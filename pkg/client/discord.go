package client

import (
	"context"

	"github.com/kylrth/disco-bouncer/internal/server"
)

// DiscordService is used to perform actions on Discord through the bouncerbot.
type DiscordService struct {
	c *Client
}

// Migrate moves an existing Discord user from the pre-ACME role to the cohort role specified by the
// given year.
func (s *DiscordService) Migrate(ctx context.Context, name, year string) error {
	const p = "/api/discord/migrate"

	m := server.Migration{
		Name: name,
		Year: year,
	}

	res, err := s.c.postJSON(ctx, p, m)
	if err == nil {
		res.Body.Close() // if it was 200 OK, the response body is empty
	}

	return err
}
