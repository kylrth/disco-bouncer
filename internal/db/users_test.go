package db

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cobaltspeech/log/pkg/testinglog"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v2"
)

func TestUserTable(t *testing.T) {
	t.Parallel()

	mockDB, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("error opening mock db: %v", err)
	}
	defer mockDB.Close()

	logger := testinglog.NewConvenientLogger(t)
	table := &UserTable{logger, mockDB}
	ctx := context.Background()

	john := User{
		Name:        "John Doe",
		FinishYear:  2019,
		AlumniBoard: true,
	}
	stephen := User{
		Name:       "Stephen Wolfram",
		FinishYear: 0,
		Professor:  true,
	}

	// create user John and check data
	mockDB.ExpectQuery("INSERT INTO users").
		WithArgs(
			john.Name, john.FinishYear, john.Professor, john.TA, john.StudentLeadership,
			john.AlumniBoard,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
	john.ID, err = table.CreateUser(ctx, &john)
	if err != nil {
		t.Errorf("error from CreateUser: %v", err)
	}
	if john.ID != 1 {
		t.Errorf("wrong ID for John: %d", john.ID)
	}

	mockDB.ExpectQuery("SELECT " + userFields + " FROM users").
		WithArgs(1).
		WillReturnRows(pgxmock.NewRows(strings.Split(userFields, ", ")).
			AddRow(
				john.Name, john.FinishYear, john.Professor, john.TA, john.StudentLeadership,
				john.AlumniBoard,
			),
		)
	newJohn, err := table.GetUser(ctx, john.ID)
	if err != nil {
		t.Errorf("error from GetUser: %v", err)
	}
	diff := cmp.Diff(&john, newJohn)
	if diff != "" {
		t.Error("unexpected John info (-want +got):\n" + diff)
	}

	// get nonexistent user
	mockDB.ExpectQuery("SELECT " + userFields + " FROM users").
		WithArgs(2).
		WillReturnError(pgx.ErrNoRows)
	_, err = table.GetUser(ctx, 2)
	if !errors.Is(err, ErrNoUser) {
		t.Errorf("unexpected error from GetUser: %v", err)
	}

	// add Stephen and get all users
	mockDB.ExpectQuery("INSERT INTO users").
		WithArgs(
			stephen.Name, stephen.FinishYear, stephen.Professor, stephen.TA,
			stephen.StudentLeadership, stephen.AlumniBoard,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(2))
	stephen.ID, err = table.CreateUser(ctx, &stephen)
	if err != nil {
		t.Errorf("error from CreateUser: %v", err)
	}
	if stephen.ID != 2 {
		t.Errorf("wrong ID for Stephen: %d", stephen.ID)
	}

	mockDB.ExpectQuery("SELECT id, " + userFields + " FROM users").
		WillReturnRows(pgxmock.NewRows(append([]string{"id"}, strings.Split(userFields, ", ")...)).
			AddRows(
				[]any{
					john.ID, john.Name, john.FinishYear, john.Professor, john.TA,
					john.StudentLeadership, john.AlumniBoard,
				},
				[]any{
					stephen.ID, stephen.Name, stephen.FinishYear, stephen.Professor, stephen.TA,
					stephen.StudentLeadership, stephen.AlumniBoard,
				},
			),
		)
	users, err := table.GetUsers(ctx)
	if err != nil {
		t.Errorf("unexpected error from GetUsers: %v", err)
	}
	diff = cmp.Diff([]*User{&john, &stephen}, users)
	if diff != "" {
		t.Error("unexpected users (-want +got):\n" + diff)
	}

	// modify Stephen and check
	stephen.Name = "Stephen King"
	mockDB.ExpectExec("UPDATE users").
		WithArgs(
			stephen.ID, stephen.Name, stephen.FinishYear, stephen.Professor, stephen.TA,
			stephen.StudentLeadership, stephen.AlumniBoard,
		).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	err = table.UpdateUser(ctx, &stephen)
	if err != nil {
		t.Errorf("unexpected error from UpdateUser: %v", err)
	}

	mockDB.ExpectQuery("SELECT " + userFields + " FROM users").
		WithArgs(stephen.ID).
		WillReturnRows(pgxmock.NewRows(strings.Split(userFields, ", ")).
			AddRow(
				stephen.Name, stephen.FinishYear, stephen.Professor, stephen.TA,
				stephen.StudentLeadership, stephen.AlumniBoard,
			),
		)
	newUser, err := table.GetUser(ctx, stephen.ID)
	if err != nil {
		t.Errorf("unexpected error from GetUser: %v", err)
	}
	diff = cmp.Diff(&stephen, newUser)
	if diff != "" {
		t.Error("unexpected Stephen info (-want +got):\n" + diff)
	}

	// update nonexistent user
	stephen.ID = 4
	mockDB.ExpectExec("UPDATE users").
		WithArgs(
			stephen.ID, stephen.Name, stephen.FinishYear, stephen.Professor, stephen.TA,
			stephen.StudentLeadership, stephen.AlumniBoard,
		).WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	err = table.UpdateUser(ctx, &stephen)
	if !errors.Is(err, ErrNoUser) {
		t.Errorf("unexpected error from UpdateUser: %v", err)
	}
	stephen.ID = 2

	// delete nonexistent user
	mockDB.ExpectExec("DELETE FROM users").
		WithArgs(4).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	err = table.DeleteUser(ctx, 4)
	if !errors.Is(err, ErrNoUser) {
		t.Errorf("unexpected error from DeleteUser: %v", err)
	}

	// delete John and check
	mockDB.ExpectExec("DELETE FROM users").
		WithArgs(john.ID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	err = table.DeleteUser(ctx, john.ID)
	if err != nil {
		t.Errorf("unexpected error from DeleteUser: %v", err)
	}

	mockDB.ExpectQuery("SELECT " + userFields + " FROM users").
		WithArgs(john.ID).
		WillReturnError(pgx.ErrNoRows)
	_, err = table.GetUser(ctx, john.ID)
	if !errors.Is(err, ErrNoUser) {
		t.Errorf("unexpected error from GetUser: %v", err)
	}

	if err = mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled DB expectations: %v", err)
	}
	logger.Done()
}
