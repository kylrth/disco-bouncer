package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cobaltspeech/log"
	"github.com/kylrth/disco-bouncer/pkg/client"
	"github.com/spf13/cobra"
)

var changePassCmd = &cobra.Command{
	Use:   "changepass",
	Short: "Change your password",
	Args:  cobra.NoArgs,
	Run:   withLAndC(changePass),
}

func changePass(_ log.Logger, c *client.Client, _ []string) error {
	// prompt for password
	fmt.Fprint(os.Stderr, "New password: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}

		return errors.New("no password entered")
	}
	password := scanner.Text()

	return c.Admin.ChangePassword(context.Background(), os.Getenv("BOUNCER_PASS"), password)
}
