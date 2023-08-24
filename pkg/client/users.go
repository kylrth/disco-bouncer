package client

import (
	"context"
	"net/url"
	"strconv"

	"github.com/kylrth/disco-bouncer/internal/db"
)

// UserService is used to view and modify the users table on the server.
type UsersService struct {
	c *Client
}

// GetAllUsers gets the current users table.
func (s *UsersService) GetAllUsers(ctx context.Context) ([]*db.User, error) {
	const p = "/api/users"

	var out []*db.User

	err := s.c.getJSON(ctx, p, &out)

	return out, err
}

// GetUser gets the information of the user with the specified ID.
func (s *UsersService) GetUser(ctx context.Context, id int) (*db.User, error) {
	p, err := url.JoinPath("/api/users", strconv.Itoa(id))
	if err != nil {
		return nil, err
	}

	var out db.User

	return &out, s.c.getJSON(ctx, p, &out)
}

// CreateUser creates a new user and returns the ID. u.ID is ignored.
func (s *UsersService) CreateUser(ctx context.Context, u *db.User) (int, error) {
	const p = "/api/users"

	var out db.User

	return out.ID, s.c.postJSONrecvJSON(ctx, p, u, &out)
}

// UpdateUser updates the information for an existing user, selected by u.ID.
func (s *UsersService) UpdateUser(ctx context.Context, u *db.User) error {
	p, err := url.JoinPath("/api/users", strconv.Itoa(u.ID))
	if err != nil {
		return err
	}

	var out db.User

	return s.c.putJSONrecvJSON(ctx, p, u, &out)
}

// DeleteUser removes the user from the server.
func (s *UsersService) DeleteUser(ctx context.Context, id int) error {
	p, err := url.JoinPath("/api/users", strconv.Itoa(id))
	if err != nil {
		return err
	}

	return s.c.delete(ctx, p)
}
