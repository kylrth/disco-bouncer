package client

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// AdminService is used to manage authentication with the server.
type AdminService struct {
	c *Client
}

var (
	// ErrInvalidCredentials is returned when the server rejects credentials while logging in.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrPasswordTooShort is returned by ChangePassword if the password is too short.
	ErrPasswordTooShort = errors.New("password too short")

	// ErrPasswordTooLong is returned by ChangePassword if the password is too long.
	ErrPasswordTooLong = errors.New("password too long")
)

// Login and store the session for later use by the client.
func (s *AdminService) Login(ctx context.Context, user, pass string) error {
	const p = "/login"

	body := map[string]string{
		"username": user,
		"password": pass,
	}

	resp, err := s.c.postJSON(ctx, p, body)
	if err != nil {
		if errors.Is(err, ErrNotLoggedIn) {
			return ErrInvalidCredentials
		}

		return err
	}
	resp.Body.Close() // If it was 200 OK, the body is "Login successful".

	return nil
}

// Logout invalidates the currently held session. You will need to call Login again in order to use
// the API.
func (s *AdminService) Logout(ctx context.Context) error {
	const p = "/logout"

	resp, err := s.c.post(ctx, p, "", nil)
	if err != nil {
		return err
	}
	resp.Body.Close() // If it was 200 OK, the body is "Logout successful".

	return nil
}

// ChangePassword updates the password on the server. Must be logged in.
func (s *AdminService) ChangePassword(ctx context.Context, pass string) error {
	const p = "/admin/pass"

	body := map[string]string{
		"password": pass,
	}

	resp, err := s.c.postJSON(ctx, p, body)
	if err != nil {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			if strings.Contains(err.Error(), "Invalid session data") {
				return ErrNotLoggedIn
			}
			if strings.Contains(err.Error(), "Password too short") {
				return ErrPasswordTooShort
			}
			if strings.Contains(err.Error(), "Password too long") {
				return ErrPasswordTooLong
			}

			fallthrough
		default:
			return err
		}
	}
	resp.Body.Close() // If it was 200 OK, the body is "Password updated successfully"

	return nil
}
