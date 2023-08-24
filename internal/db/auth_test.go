package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/cobaltspeech/log/pkg/testinglog"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v2"
)

// AnyBcrypt satisfies pgxmock.Argument
type AnyBcrypt struct{}

// Match matches any string that starts with "$2a$10$"
func (a AnyBcrypt) Match(v interface{}) bool {
	switch val := v.(type) {
	case string:
		fmt.Println(val)
		return strings.HasPrefix(val, "$2a$10$")
	default:
		return false
	}
}

func TestAdminTable(t *testing.T) {
	t.Parallel()

	mockDB, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("error opening mock db: %v", err)
	}
	defer mockDB.Close()

	logger := testinglog.NewConvenientLogger(t)
	table := &AdminTable{logger: logger, pool: mockDB}
	ctx := context.Background()

	// create admin John and check password
	mockDB.ExpectExec("INSERT INTO admins").
		WithArgs("john", AnyBcrypt{}).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	err = table.AddAdmin(ctx, "john", "doe")
	if err != nil {
		t.Errorf("error from AddAdmin: %v", err)
	}

	mockDB.ExpectQuery("SELECT password FROM admins").
		WithArgs("john").
		WillReturnRows(pgxmock.NewRows([]string{"password"}).
			AddRow("$2a$10$mptpjWROiQEfYu2l1Og4y.o7yssk4y8mnQh2.vS1KXufsb5MW3Wf."),
		)
	success, err := table.CheckPassword(ctx, "john", "doe")
	if err != nil {
		t.Errorf("error from CheckPassword: %v", err)
	}
	if !success {
		t.Errorf("password check for john failed")
	}

	// check nonexistent Stephen's password
	mockDB.ExpectQuery("SELECT password FROM admins").
		WithArgs("stephen").
		WillReturnError(pgx.ErrNoRows)
	success, err = table.CheckPassword(ctx, "stephen", "doe")
	if err != nil {
		t.Errorf("error from CheckPassword: %v", err)
	}
	if success {
		t.Errorf("password check for stephen passed")
	}

	// change John's password and check it
	mockDB.ExpectExec("UPDATE admins").
		WithArgs("john", AnyBcrypt{}).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	err = table.ChangePassword(ctx, "john", "password")
	if err != nil {
		t.Errorf("error from ChangePassword: %v", err)
	}

	mockDB.ExpectQuery("SELECT password FROM admins").
		WithArgs("john").
		WillReturnRows(pgxmock.NewRows([]string{"password"}).
			AddRow("$2a$10$4mHLLmN4VRKUnf0filfiT.d5I3qycV4tSRoKKIWirKXK52pVU1oWW"),
		)
	success, err = table.CheckPassword(ctx, "john", "doe")
	if err != nil {
		t.Errorf("error from CheckPassword: %v", err)
	}
	if success {
		t.Errorf("password check for john passed with old password")
	}

	mockDB.ExpectQuery("SELECT password FROM admins").
		WithArgs("john").
		WillReturnRows(pgxmock.NewRows([]string{"password"}).
			AddRow("$2a$10$4mHLLmN4VRKUnf0filfiT.d5I3qycV4tSRoKKIWirKXK52pVU1oWW"),
		)
	success, err = table.CheckPassword(ctx, "john", "password")
	if err != nil {
		t.Errorf("error from CheckPassword: %v", err)
	}
	if !success {
		t.Errorf("password check for john failed")
	}

	// change Stephen's password (fail)
	mockDB.ExpectExec("UPDATE admins").
		WithArgs("stephen", AnyBcrypt{}).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	err = table.ChangePassword(ctx, "stephen", "password")
	if !errors.Is(err, ErrNoUser) {
		t.Errorf("unexpected error from ChangePassword: %v", err)
	}

	// delete Stephen's account (fail)
	mockDB.ExpectExec("DELETE FROM admins").
		WithArgs("stephen").
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	err = table.DeleteAdmin(ctx, "stephen")
	if !errors.Is(err, ErrNoUser) {
		t.Errorf("unexpected error from DeleteAdmin: %v", err)
	}

	// delete John's account
	mockDB.ExpectExec("DELETE FROM admins").
		WithArgs("john").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	err = table.DeleteAdmin(ctx, "john")
	if err != nil {
		t.Errorf("unexpected error from DeleteAdmin: %v", err)
	}

	if err = mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled DB expectations: %v", err)
	}
	logger.Done()
}
