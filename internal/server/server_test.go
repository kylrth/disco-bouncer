// Package server_test defines integration tests for the server.
package server_test

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/cobaltspeech/log"
	"github.com/cobaltspeech/log/pkg/testinglog"
	"github.com/gofiber/fiber/v2"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kylrth/disco-bouncer/internal/db"
	"github.com/kylrth/disco-bouncer/internal/server"
	"github.com/kylrth/disco-bouncer/pkg/bouncerbot"
	"github.com/kylrth/disco-bouncer/pkg/client"
	"github.com/kylrth/disco-bouncer/pkg/encrypt"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func TestMain(m *testing.M) {
	flag.Parse() // must be run before calling testing.Short()
	if testing.Short() {
		fmt.Println("skipping integration tests")

		return
	}

	done, err := setupDBPool()
	if err != nil {
		fmt.Println("failed to set up database pool:", err)
		done()
		os.Exit(1)
	}

	code := m.Run()

	done() // This can't be deferred because os.Exit won't run deferred code.

	os.Exit(code)
}

var dbPool *pgxpool.Pool

func setupDBPool() (done func(), err error) {
	done = func() {}
	// connect to Docker
	pool, err := dockertest.NewPool("")
	if err != nil {
		return done, fmt.Errorf("could not construct docker pool: %w", err)
	}
	err = pool.Client.Ping()
	if err != nil {
		return done, fmt.Errorf("could not connect to docker: %w", err)
	}

	const pgUser = "testuser"
	const pgPass = "testpass"
	const hardTimeout = 120 // seconds until hard shutdown

	// start Postgres container
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "18",
		Env: []string{
			"POSTGRES_USER=" + pgUser,
			"POSTGRES_PASSWORD=" + pgPass,
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.NeverRestart()
	})
	if err != nil {
		return done, fmt.Errorf("could not start postgres instance: %w", err)
	}
	done = func() {
		finalErr := pool.Purge(resource)
		if finalErr != nil {
			fmt.Println("could not purge postgres instance:", finalErr)
		}
	}
	err = resource.Expire(hardTimeout)
	if err != nil {
		return done, fmt.Errorf("could not set expiration on postgres instance: %w", err)
	}

	// connect to Postgres
	pool.MaxWait = hardTimeout * time.Second
	pgURI := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		pgUser, pgPass, resource.GetHostPort("5432/tcp"), pgUser)
	err = pool.Retry(func() error {
		// apply migrations to set up tables
		newErr := db.ApplyMigrations(pgURI)
		if newErr != nil && !isStartupErr(newErr) {
			return backoff.Permanent(newErr)
		}

		return newErr
	})
	if err != nil {
		return done, fmt.Errorf("could not connect to postgres instance: %w", err)
	}

	// open connection
	dbPool, err = pgxpool.New(context.Background(), pgURI)
	if err != nil {
		return done, fmt.Errorf("failed to connect after running migrations: %w", err)
	}

	return done, nil
}

var startupErrors = []string{
	"connection reset by peer",
	"the database system is starting up",
	"connect: EOF",
}

func isStartupErr(err error) bool {
	s := err.Error()

	for _, search := range startupErrors {
		if strings.Contains(s, search) {
			return true
		}
	}

	return false
}

const (
	testUser = "test"
	testPass = "testtest"
	addr     = ":8321"
)

func setupServer(t *testing.T, l log.Logger) (done func()) {
	t.Helper()

	aTable := db.NewAdminTable(l, dbPool)
	uTable := db.NewUserTable(l, dbPool)

	// create a new admin user
	err := aTable.AddAdmin(context.Background(), testUser, testPass)
	if err != nil {
		t.Fatalf("failed to create admin user: %v", err)
	}

	// start server
	app := fiber.New()
	server.AddAuthHandlers(l, app, dbPool, aTable)
	server.AddCRUDHandlers(l, app, uTable)

	go func() {
		serveErr := app.Listen(addr)
		if serveErr != nil {
			l.Error("msg", "error from server listener", "error", serveErr)
		}
	}()

	// wait for the server to accept connections
	time.Sleep(time.Second / 2)

	// Define a function to shut down the server and clean up the database.
	return func() {
		finalErr := app.ShutdownWithTimeout(5 * time.Second)
		if finalErr != nil {
			t.Errorf("error shutting down server: %v", finalErr)
		}

		users, finalErr := uTable.GetUsers(context.Background())
		if finalErr != nil {
			t.Errorf("error getting users to delete: %v", finalErr)
		} else {
			for _, u := range users {
				finalErr = uTable.DeleteUser(context.Background(), u.ID)
				if finalErr != nil {
					t.Errorf("error deleting user %d: %v", u.ID, finalErr)
				}
			}
		}

		finalErr = aTable.DeleteAdmin(context.Background(), testUser)
		if finalErr != nil {
			t.Errorf("error deleting admin user: %v", finalErr)
		}
	}
}

//nolint:paralleltest // This test uses a database.
func TestAll(t *testing.T) { //nolint:cyclop,funlen,gocyclo // long integration test
	l := testinglog.NewConvenientLogger(t, testinglog.WithFieldIgnoreFunc(ignoreIDs))
	t.Cleanup(l.Done)

	ctx := context.Background()

	shutdown := setupServer(t, l)
	t.Cleanup(shutdown)

	// define client
	c, err := client.NewClient("http://localhost" + addr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = c.Admin.Login(ctx, testUser, testPass)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	newPass := "newPass1"
	err = c.Admin.ChangePassword(ctx, testPass, newPass)
	if err != nil {
		t.Fatalf("failed to change password: %v", err)
	}

	err = c.Admin.Logout(ctx)
	if err != nil {
		t.Fatalf("failed to logout: %v", err)
	}

	err = c.Admin.Login(ctx, testUser, newPass)
	if err != nil {
		t.Fatalf("failed to login with new password: %v", err)
	}

	users, err := c.Users.GetAllUsers(ctx)
	if err != nil {
		t.Fatalf("failed to get users: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected no users, got %d", len(users))
	}

	u1 := db.User{
		Name:        "John Doe",
		NameKeyHash: "asdfjkl",
		FinishYear:  "2021",
	}
	u2 := db.User{
		Name:        "Jason Mendoza",
		NameKeyHash: "lkjfdsa",
		FinishYear:  "2019",
		AlumniBoard: true,
	}
	u1.ID, err = c.Users.CreateUser(ctx, &u1)
	if err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	u2.ID, err = c.Users.CreateUser(ctx, &u2)
	if err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	u1.TA = true
	err = c.Users.UpdateUser(ctx, &u1)
	if err != nil {
		t.Fatalf("failed to update user1: %v", err)
	}

	users, err = c.Users.GetAllUsers(ctx)
	if err != nil {
		t.Errorf("failed to get users: %v", err)
	}
	diff := cmp.Diff(
		[]*db.User{&u1, &u2}, users,
		cmpopts.SortSlices(func(x, y *db.User) bool { return x.ID < y.ID }),
	)
	if diff != "" {
		t.Error("unexpected users (-want +got):\n" + diff)
	}

	users, err = c.Users.GetAllUsers(ctx, client.WithKeyHash(u1.NameKeyHash))
	if err != nil {
		t.Errorf("failed to get filtered users: %v", err)
	}
	if diff = cmp.Diff([]*db.User{&u1}, users); diff != "" {
		t.Error("unexpected users (-want +got):\n" + diff)
	}

	err = c.Users.DeleteUser(ctx, u2.ID)
	if err != nil {
		t.Fatalf("failed to delete user2: %v", err)
	}

	newU1, err := c.Users.GetUser(ctx, u1.ID)
	if err != nil {
		t.Errorf("failed to get user: %v", err)
	}
	if diff := cmp.Diff(&u1, newU1); diff != "" {
		t.Error("unexpected users (-want +got):\n" + diff)
	}

	users, err = c.Users.GetAllUsers(ctx)
	if err != nil {
		t.Errorf("failed to get users: %v", err)
	}
	if diff := cmp.Diff([]*db.User{&u1}, users); diff != "" {
		t.Error("unexpected users (-want +got):\n" + diff)
	}
}

//nolint:paralleltest // This test uses a database.
func TestUploadAndDecrypt(t *testing.T) { //nolint:cyclop,funlen,gocyclo // long integration test
	l := testinglog.NewConvenientLogger(t, testinglog.WithFieldIgnoreFunc(ignoreKeyHash))
	t.Cleanup(l.Done)

	ctx := context.Background()

	shutdown := setupServer(t, l)
	t.Cleanup(shutdown)

	// define client
	c, err := client.NewClient("http://localhost" + addr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = c.Admin.Login(ctx, testUser, testPass)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	const u1Name = "John Doe"
	const u2Name = "Jason Mendoza"
	u1 := db.User{
		Name:       u1Name,
		FinishYear: "2021",
	}
	u2 := db.User{
		Name:        u2Name,
		FinishYear:  "2019",
		AlumniBoard: true,
	}
	var u1Key, u2Key string

	// upload the users
	u1.ID, u1Key, err = c.Users.Upload(ctx, &u1)
	if err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	u2.ID, u2Key, err = c.Users.Upload(ctx, &u2)
	if err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	// hash the generated keys
	u1.NameKeyHash, err = encrypt.MD5Hash(u1Key)
	if err != nil {
		t.Fatalf("failed to hash key: %v", err)
	}
	u2.NameKeyHash, err = encrypt.MD5Hash(u2Key)
	if err != nil {
		t.Fatalf("failed to hash key: %v", err)
	}

	// get by key hash
	users, err := c.Users.GetAllUsers(ctx, client.WithKeyHash(u1.NameKeyHash))
	if err != nil {
		t.Errorf("failed to get filtered users: %v", err)
	}
	if diff := cmp.Diff([]*db.User{&u1}, users); diff != "" {
		t.Error("unexpected users (-want +got):\n" + diff)
	}
	users, err = c.Users.GetAllUsers(ctx, client.WithKeyHash(u2.NameKeyHash))
	if err != nil {
		t.Errorf("failed to get filtered users: %v", err)
	}
	if diff := cmp.Diff([]*db.User{&u2}, users); diff != "" {
		t.Error("unexpected users (-want +got):\n" + diff)
	}

	// decrypt on the server side
	dec := bouncerbot.TableDecrypter{Table: db.NewUserTable(l, dbPool)}

	out, err := dec.Decrypt(u1Key)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}
	u1.Name = u1Name // expect the unecrypted name
	if diff := cmp.Diff(&u1, out); diff != "" {
		t.Error("unexpected decrypted info (-want +got):\n" + diff)
	}
	err = dec.Delete(u1.ID)
	if err != nil {
		t.Errorf("unexpected error from Decrypter.Delete: %v", err)
	}

	// (try again, should get ErrNotFound)
	_, err = dec.Decrypt(u1Key)
	if !errors.Is(err, bouncerbot.ErrNotFound) {
		t.Errorf("expected error decrypting u1 again, got %v", err)
	}

	out, err = dec.Decrypt(u2Key)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}
	u2.Name = u2Name // expect the unecrypted name
	if diff := cmp.Diff(&u2, out); diff != "" {
		t.Error("unexpected decrypted info (-want +got):\n" + diff)
	}
	err = dec.Delete(u2.ID)
	if err != nil {
		t.Errorf("unexpected error from Decrypter.Delete: %v", err)
	}

	// now it should be empty
	users, err = c.Users.GetAllUsers(ctx)
	if err != nil {
		t.Fatalf("failed to get users: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected no users, got %d", len(users))
	}
}

func ignoreIDs(map[string]string) []string {
	return []string{"id"}
}

func ignoreKeyHash(fields map[string]string) []string {
	out := ignoreIDs(fields)

	msg, ok := fields["msg"]
	if !ok || msg != "got all users" {
		return out
	}

	_, ok = fields["keyHash"]
	if ok {
		return append(out, "keyHash")
	}

	return out
}
