package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/kylrth/disco-bouncer/internal/db"
	"github.com/kylrth/disco-bouncer/pkg/encrypt"
)

// UsersService is used to view and modify the users table on the server.
type UsersService struct {
	c *Client
}

type filters struct {
	keyHash string
}

// FilterOption is a way to add filter options to the request when calling GetAllUsers.
type FilterOption = func(f *filters)

// WithKeyHash returns a FilterOption that filters by the provided MD5 key hash.
func WithKeyHash(keyHash string) FilterOption {
	return func(f *filters) { f.keyHash = keyHash }
}

func (f *filters) formatQueryParams() string {
	var out strings.Builder

	if f.keyHash != "" {
		out.WriteString(",keyHash=" + f.keyHash)
	}

	if out.Len() == 0 {
		return ""
	}

	return "?" + out.String()[1:] // remove initial ","
}

// GetAllUsers gets the current users table.
func (s *UsersService) GetAllUsers(ctx context.Context, opts ...FilterOption) ([]*db.User, error) {
	var f filters
	for _, opt := range opts {
		opt(&f)
	}

	p := "/api/users" + f.formatQueryParams()

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

// Upload uploads a new user to the server. It encrypts u.Name, fills in u.NameKeyHash, and returns
// the received ID and the encrypted key. The fields of u will be updated.
func (s *UsersService) Upload(ctx context.Context, u *db.User) (id int, key string, err error) {
	u.Name, key, err = encrypt.Encrypt(u.Name)
	if err != nil {
		return 0, key, fmt.Errorf("encrypt name: %w", err)
	}
	u.NameKeyHash, err = encrypt.MD5Hash(key)
	if err != nil {
		return 0, key, fmt.Errorf("hash key: %w", err)
	}

	userID, err := s.CreateUser(ctx, u)

	return userID, key, err
}
