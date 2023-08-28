package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cobaltspeech/log"
	"github.com/kylrth/disco-bouncer/pkg/client"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete ID [IDS...]",
	Short: "Delete users by ID",
	Args:  cobra.MinimumNArgs(1),
	Run:   withLAndC(deleteIDs),
}

func deleteIDs(l log.Logger, c *client.Client, ids []string) error {
	for _, id := range ids {
		i, err := strconv.Atoi(id)
		if err != nil {
			l.Error("msg", "invalid ID", "id", id, "error", err)

			continue
		}

		err = c.Users.DeleteUser(context.Background(), i)
		if err != nil {
			return fmt.Errorf("delete user %d: %w", i, err)
		}
	}

	return nil
}
