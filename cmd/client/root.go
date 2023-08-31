package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cobaltspeech/log"
	"github.com/cobaltspeech/log/pkg/level"
	"github.com/kylrth/disco-bouncer/pkg/client"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
}

var rootCmd = &cobra.Command{
	Short: "Avoid Panic! at the Disco(rd)",
}

var (
	verbosity int
	serverURL string
)

func init() {
	rootCmd.AddCommand(
		uploadCmd,
		getCmd,
		deleteCmd,
		changePassCmd,
	)

	rootCmd.PersistentFlags().IntVarP(&verbosity, "verbosity", "v", 2, "set verbosity (1-4)")
	rootCmd.PersistentFlags().StringVarP(
		&serverURL, "server", "s", "http://localhost:8321", "server URL")
}

func withLogger(f func(log.Logger, []string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		l := log.NewLeveledLogger(log.WithFilterLevel(level.Verbosity(verbosity)))

		err := f(l, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
}

func withLAndC(f func(log.Logger, *client.Client, []string) error) func(*cobra.Command, []string) {
	return withLogger(func(l log.Logger, args []string) error {
		c, err := client.NewClient(serverURL)
		if err != nil {
			return err
		}

		err = c.Admin.Login(
			context.Background(), os.Getenv("BOUNCER_USER"), os.Getenv("BOUNCER_PASS"))
		if err != nil {
			_, userSet := os.LookupEnv("BOUNCER_USER")
			_, passSet := os.LookupEnv("BOUNCER_PASS")
			if !userSet || !passSet {
				fmt.Fprintln(os.Stderr,
					"Be sure to set credentials with BOUNCER_USER and BOUNCER_PASS")
			}

			return fmt.Errorf("login: %w", err)
		}
		defer func() {
			err := c.Admin.Logout(context.Background())
			if err != nil {
				l.Error("msg", "failed to log out", "error", err)
			}
		}()

		return f(l, c, args)
	})
}
