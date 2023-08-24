// Package server_test defines integration tests for the server.
package server_test

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/cobaltspeech/log/pkg/testinglog"
	"github.com/gofiber/fiber/v2"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kylrth/disco-bouncer/internal/db"
	"github.com/kylrth/disco-bouncer/internal/server"
	"github.com/kylrth/disco-bouncer/pkg/client"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var dbPool *pgxpool.Pool

func TestMain(m *testing.M) { //nolint:cyclop // lots of setup
	flag.Parse() // must be run before calling testing.Short()
	if testing.Short() {
		fmt.Println("skipping integration tests")

		return
	}

	// connect to Docker
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not construct docker pool: %v", err)
	}
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("could not connect to docker: %v", err)
	}

	const pgUser = "testuser"
	const pgPass = "testpass"
	const hardTimeout = 120 // seconds until hard shutdown

	// start Postgres container
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13",
		Env: []string{
			"POSTGRES_USER=" + pgUser,
			"POSTGRES_PASSWORD=" + pgPass,
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.NeverRestart()
	})
	if err != nil {
		log.Fatalf("could not start postgres instance: %v", err)
	}
	err = resource.Expire(hardTimeout)
	if err != nil {
		log.Fatalf("could not set expiration on postgres instance: %v", err)
	}

	// connect to Postgres
	pool.MaxWait = hardTimeout * time.Second
	pgURI := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		pgUser, pgPass, resource.GetHostPort("5432/tcp"), pgUser)
	if err = pool.Retry(func() error {
		// apply migrations to set up tables
		newErr := db.ApplyMigrations(pgURI)
		if newErr != nil && !strings.Contains(newErr.Error(), "connection reset by peer") {
			return backoff.Permanent(newErr)
		}

		return newErr
	}); err != nil {
		log.Fatalf("could not connect to postgres instance: %v", err)
	}
	// open connection
	dbPool, err = pgxpool.New(context.Background(), pgURI)
	if err != nil {
		log.Fatalf("failed to connect after running migrations: %v", err)
	}

	// run tests
	code := m.Run()

	// This can't be deferred because os.Exit won't run deferred code.
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge postgres instance: %v", err)
	}

	os.Exit(code)
}

//nolint:paralleltest // This test uses a database.
func TestAll(t *testing.T) { //nolint:cyclop,funlen,gocyclo // testing sequential operations
	l := testinglog.NewConvenientLogger(t)
	defer l.Done()

	const testUser = "test"
	const testPass = "testtest"

	ctx := context.Background()

	// create a new admin user
	err := db.NewAdminTable(l, dbPool).AddAdmin(ctx, testUser, testPass)
	if err != nil {
		t.Fatalf("failed to create admin user: %v", err)
	}

	// start server
	app := fiber.New()
	server.AddAuthHandlers(l, app, dbPool)
	server.AddCRUDHandlers(l, app, dbPool)

	const addr = ":8321"
	go func() {
		serveErr := app.Listen(addr)
		if serveErr != nil {
			l.Error("msg", "error from server listener", "error", serveErr)
		}
	}()
	defer func() {
		shutdownErr := app.ShutdownWithTimeout(5 * time.Second)
		if shutdownErr != nil {
			t.Errorf("error shutting down server: %v", shutdownErr)
		}
	}()

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
	err = c.Admin.ChangePassword(ctx, newPass)
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
		Name:       "John Doe",
		FinishYear: 2021,
	}
	u2 := db.User{
		Name:        "Jason Mendoza",
		FinishYear:  2019,
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
