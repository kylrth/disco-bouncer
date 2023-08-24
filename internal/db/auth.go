package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/cobaltspeech/log"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// AdminTable represents the table of admins permitted to use the service. It handles password
// hashing, so all `pass` method arguments are expected to be plaintext.
type AdminTable struct {
	logger log.Logger
	pool   PgxIface
}

// NewAdminTable creates a new AdminTable backed by a Postgres connection pool.
func NewAdminTable(l log.Logger, pool PgxIface) *AdminTable {
	out := AdminTable{
		logger: l,
		pool:   pool,
	}

	return &out
}

// AddAdmin creates a new admin account.
func (a *AdminTable) AddAdmin(ctx context.Context, user, pass string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		a.logger.Error("msg", "failed to hash password for new admin", "user", user, "error", err)

		return fmt.Errorf("hash password: %w", err)
	}

	_, err = a.pool.Exec(
		ctx, "INSERT INTO admins (username, password) VALUES ($1, $2)", user, string(hashed),
	)
	if err != nil {
		a.logger.Error("msg", "failed to store new admin", "user", user, "error", err)

		return err
	}

	a.logger.Debug("msg", "stored new admin", "user", user)

	return nil
}

// DeleteAdmin removes an admin account. It does not affect any changes they may have made to the
// data.
func (a *AdminTable) DeleteAdmin(ctx context.Context, user string) error {
	tag, err := a.pool.Exec(ctx, "DELETE FROM admins WHERE username=$1", user)
	if err != nil {
		a.logger.Error("msg", "failed to delete admin", "user", user, "error", err)

		return err
	}
	if tag.RowsAffected() < 1 {
		a.logger.Info("msg", "no user found to delete", "user", user)

		return ErrNoUser
	}

	a.logger.Debug("msg", "deleted admin", "user", user)

	return nil
}

// CheckPassword returns true if the admin exists and the password matches the hash on file.
// Otherwise it returns false.
func (a *AdminTable) CheckPassword(ctx context.Context, user, pass string) (bool, error) {
	var hashed string
	err := a.pool.QueryRow(ctx, "SELECT password FROM admins WHERE username=$1", user).
		Scan(&hashed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			a.logger.Info("msg", "checked password for nonexistent user", "user", user)

			return false, nil
		}

		a.logger.Error("msg", "failed to check password", "user", user, "error", err)

		return false, err
	}

	passed := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(pass)) == nil

	if passed {
		a.logger.Debug("msg", "successful password check", "user", user)
	} else {
		a.logger.Debug("msg", "unsuccessful password check", "user", user)
	}

	return passed, nil
}

// ErrNoUser is returned by if the user or admin account is not found.
var ErrNoUser = errors.New("user not found")

// ChangePassword updates the hashed password for an admin. ErrNoUser is returned if the admin is
// not in the database.
func (a *AdminTable) ChangePassword(ctx context.Context, user, pass string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		a.logger.Error("msg", "failed to hash new password", "user", user, "error", err)

		return err
	}

	tag, err := a.pool.Exec(
		ctx, "UPDATE admins SET password=$2 WHERE username=$1", user, string(hashed))
	if err != nil {
		a.logger.Error("msg", "failed to update password", "user", user, "error", err)

		return err
	}
	if tag.RowsAffected() != 1 {
		a.logger.Info("msg", "no user found to update password", "user", user)

		return ErrNoUser
	}

	a.logger.Debug("msg", "password updated", "user", user)

	return nil
}
