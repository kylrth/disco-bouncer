// Package server_test defines integration tests for the server.
package server_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/cobaltspeech/log/pkg/testinglog"
	"github.com/gofiber/fiber/v2"
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

func TestMain(m *testing.M) {
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
	resource.Expire(hardTimeout)

	// connect to Postgres
	pool.MaxWait = hardTimeout * time.Second
	pgURI := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		pgUser, pgPass, resource.GetHostPort("5432/tcp"), pgUser)
	if err = pool.Retry(func() error {
		// apply migrations to set up tables
		err := db.ApplyMigrations(pgURI)
		if err != nil && !strings.Contains(err.Error(), "connection reset by peer") {
			return backoff.Permanent(err)
		}
		return err
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

func TestAll(t *testing.T) {
	l := testinglog.NewConvenientLogger(t)
	defer l.Done()

	const testUser = "test"
	const testPass = "testtest"

	// create a new admin user
	db.NewAdminTable(l, dbPool).AddAdmin(context.Background(), testUser, testPass)

	const addr = ":8321"

	// start server
	app := fiber.New()
	server.AddAuthHandlers(l, app, dbPool)
	server.AddCRUDHandlers(l, app, dbPool)

	go func() {
		err := app.Listen(addr)
		if err != nil {
			l.Error("msg", "error from server listener", "error", err)
		}
	}()
	defer func() {
		err := app.ShutdownWithTimeout(5 * time.Second)
		if err != nil {
			t.Errorf("error shutting down server: %v", err)
		}
	}()

	// define client
	c, err := client.NewClient("http://localhost" + addr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = c.Admin.Login(testUser, testPass)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	newPass := "newPass1"
	err = c.Admin.ChangePassword(newPass)
	if err != nil {
		t.Fatalf("failed to change password: %v", err)
	}

	err = c.Admin.Logout()
	if err != nil {
		t.Fatalf("failed to logout: %v", err)
	}

	err = c.Admin.Login(testUser, newPass)
	if err != nil {
		t.Fatalf("failed to login with new password: %v", err)
	}

	users, err := c.Users.GetAllUsers()
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
	u1.ID, err = c.Users.CreateUser(&u1)
	if err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	u2.ID, err = c.Users.CreateUser(&u2)
	if err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	u1.TA = true
	err = c.Users.UpdateUser(&u1)
	if err != nil {
		t.Fatalf("failed to update user1: %v", err)
	}

	users, err = c.Users.GetAllUsers()
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

	err = c.Users.DeleteUser(u2.ID)
	if err != nil {
		t.Fatalf("failed to delete user2: %v", err)
	}

	newU1, err := c.Users.GetUser(u1.ID)
	if err != nil {
		t.Errorf("failed to get user: %v", err)
	}
	if diff := cmp.Diff(&u1, newU1); diff != "" {
		t.Error("unexpected users (-want +got):\n" + diff)
	}

	users, err = c.Users.GetAllUsers()
	if err != nil {
		t.Errorf("failed to get users: %v", err)
	}
	if diff := cmp.Diff([]*db.User{&u1}, users); diff != "" {
		t.Error("unexpected users (-want +got):\n" + diff)
	}
}
