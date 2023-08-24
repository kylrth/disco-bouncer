package db

import (
	"context"
	"errors"

	"github.com/cobaltspeech/log"
	"github.com/jackc/pgx/v5"
)

// UserTable represents the table of users that the bouncer will accept into the Discord server.
type UserTable struct {
	logger log.Logger
	pool   PgxIface
}

// NewUserTable creates a new UserTable backed by a Postgres connection pool.
func NewUserTable(l log.Logger, pool PgxIface) *UserTable {
	out := UserTable{
		logger: l,
		pool:   pool,
	}

	return &out
}

// User contains the information about a user necessary to admit them to the Discord server and
// assign appropriate roles upon entry to the server.
type User struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	FinishYear        int    `json:"finish_year"`
	Professor         bool   `json:"professor"`
	TA                bool   `json:"ta"`
	StudentLeadership bool   `json:"student_leadership"`
	AlumniBoard       bool   `json:"alumni_board"`
}

const (
	userFields = "name, finish_year, professor, ta, student_leadership, alumni_board"
	userSets   = "name=$2, finish_year=$3, professor=$4, ta=$5, student_leadership=$6, alumni_board=$7"
)

// GetUsers returns all users in the database.
func (t *UserTable) GetUsers(ctx context.Context) ([]*User, error) {
	rows, err := t.pool.Query(ctx, "SELECT id, "+userFields+" FROM users")
	if err != nil {
		t.logger.Error("msg", "failed to query db for users", "error", err)

		return nil, err
	}
	defer rows.Close()

	var out []*User
	for rows.Next() {
		var u User
		err = rows.Scan(
			&u.ID, &u.Name, &u.FinishYear, &u.Professor, &u.TA, &u.StudentLeadership,
			&u.AlumniBoard,
		)
		if err != nil {
			t.logger.Error("msg", "failed to scan user row", "error", err)

			return out, err
		}
		out = append(out, &u)
	}

	t.logger.Debug("msg", "got all users", "count", len(out))

	return out, nil
}

// GetUser returns the user by ID, if present. If not present, ErrNoUser is returned.
func (t *UserTable) GetUser(ctx context.Context, id int) (*User, error) {
	u := User{ID: id}
	err := t.pool.QueryRow(ctx, "SELECT "+userFields+" FROM users WHERE id=$1", id).
		Scan(&u.Name, &u.FinishYear, &u.Professor, &u.TA, &u.StudentLeadership, &u.AlumniBoard)
	if errors.Is(err, pgx.ErrNoRows) {
		t.logger.Info("msg", "user not in database", "id", id)

		return &u, ErrNoUser
	}
	if err != nil {
		t.logger.Error("msg", "failed to search for user", "id", id, "error", err)

		return &u, err
	}

	t.logger.Debug("msg", "found user info", "id", id)

	return &u, nil
}

// CreateUser creates a new user (ignoring the ID field) and returns the new ID.
func (t *UserTable) CreateUser(ctx context.Context, u *User) (int, error) {
	var newID int

	err := t.pool.QueryRow(
		ctx, "INSERT INTO users ("+userFields+") VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		u.Name, u.FinishYear, u.Professor, u.TA, u.StudentLeadership, u.AlumniBoard,
	).Scan(&newID)
	if err != nil {
		t.logger.Error("msg", "failed to create user", "error", err)

		return newID, err
	}

	t.logger.Debug("msg", "created new user", "id", newID)

	return newID, nil
}

// UpdateUser inserts the information in u into the row identified by u.ID. If that row does not
// exist, ErrNoUser is returned.
func (t *UserTable) UpdateUser(ctx context.Context, u *User) error {
	tag, err := t.pool.Exec(ctx,
		"UPDATE users SET "+userSets+" WHERE id=$1",
		u.ID, u.Name, u.FinishYear, u.Professor, u.TA, u.StudentLeadership, u.AlumniBoard,
	)
	if err != nil {
		t.logger.Error("msg", "failed to update user", "id", u.ID, "error", err)

		return err
	}
	if tag.RowsAffected() != 1 {
		t.logger.Info("msg", "no matching user to update", "id", u.ID)

		return ErrNoUser
	}

	t.logger.Debug("msg", "updated user", "id", u.ID)

	return nil
}

// DeleteUser removes the user by ID. If the user did not exist, returns ErrNoUser.
func (t *UserTable) DeleteUser(ctx context.Context, id int) error {
	tag, err := t.pool.Exec(ctx, "DELETE FROM users WHERE id=$1", id)
	if err != nil {
		t.logger.Error("msg", "failed to delete user", "id", id, "error", err)

		return err
	}
	if tag.RowsAffected() != 1 {
		t.logger.Info("msg", "no matching user to delete", "id", id)

		return ErrNoUser
	}

	t.logger.Debug("msg", "deleted user", "id", id)

	return nil
}
