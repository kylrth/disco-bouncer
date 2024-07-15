package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/cobaltspeech/log"
	"github.com/kylrth/disco-bouncer/internal/db"
	"github.com/kylrth/disco-bouncer/pkg/client"
	"github.com/kylrth/disco-bouncer/pkg/encrypt"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [ID]",
	Short: "Get user info from the server, possibly filtering by ID or key hash",
	Args: func(*cobra.Command, []string) error {
		if useHashes && useKeys {
			return errors.New("cannot use both --hashes and --keys")
		}

		return nil
	},
	Run: withLAndC(func(_ log.Logger, c *client.Client, args []string) error {
		if stdin {
			args = []string{}

			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				args = append(args, scanner.Text())
			}
		}

		return get(c, args)
	}),
}

var (
	stdin     bool
	useHashes bool
	useKeys   bool
)

func init() {
	getCmd.Flags().BoolVar(
		&stdin, "stdin", false, "read values from stdin instead of as arguments",
	)
	getCmd.Flags().BoolVar(
		&useHashes, "hashes", false, "treat arguments as key hashes to search for, instead of IDs",
	)
	getCmd.Flags().BoolVar(
		&useKeys, "keys", false, "treat arguments as keys to search with, instead of IDs. The "+
			"keys are then used to decrypt the names. The keys are *never* sent to the server.",
	)
}

func get(c *client.Client, ids []string) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// header
	w.Write([]string{ //nolint:errcheck // We're writing to stdout.
		"id", "name", "name_key_hash", "finish_year", "professor", "ta", "student_leadership",
		"alumni_board",
	})

	if len(ids) == 0 {
		// get all
		users, err := c.Users.GetAllUsers(context.Background())
		if err != nil {
			return err
		}
		writeUser(w, users...)

		return nil
	}

	if useHashes {
		// get by hashes
		for _, hash := range ids {
			users, err := c.Users.GetAllUsers(context.Background(), client.WithKeyHash(hash))
			if err != nil {
				return err
			}
			writeUser(w, users...)
		}

		return nil
	}

	if useKeys {
		return getByKeys(context.Background(), w, c, ids)
	}

	return getByIDs(w, c, ids)
}

func getByKeys(ctx context.Context, w *csv.Writer, c *client.Client, keys []string) error {
	for _, key := range keys {
		user, err := getWithKey(ctx, c, key)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				fmt.Fprintf(os.Stderr, "did not find any users encrypted with key '%s'\n", key)

				continue
			}

			return err
		}

		writeUser(w, user)
	}

	return nil
}

var ErrNotFound = errors.New("not found")

func getWithKey(ctx context.Context, c *client.Client, key string) (*db.User, error) {
	hash, err := encrypt.MD5Hash(key)
	if err != nil {
		return nil, err
	}

	users, err := c.Users.GetAllUsers(ctx, client.WithKeyHash(hash))
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		user.Name, err = encrypt.Decrypt(user.Name, key)
		if err == nil {
			return user, nil
		}

		if errors.As(err, &encrypt.ErrInauthenticated{}) {
			// wrong key (wow, hash collision!)
			continue
		}

		return nil, err
	}

	return nil, ErrNotFound
}

func getByIDs(w *csv.Writer, c *client.Client, ids []string) error {
	idInts := make([]int, 0, len(ids))
	for _, id := range ids {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			return fmt.Errorf("invalid argument: %w", err)
		}
		idInts = append(idInts, idInt)
	}

	for _, id := range idInts {
		user, err := c.Users.GetUser(context.Background(), id)
		if err != nil {
			return err
		}
		writeUser(w, user)
	}

	return nil
}

func writeUser(w *csv.Writer, us ...*db.User) {
	for _, u := range us {
		w.Write([]string{ //nolint:errcheck // We're writing to stdout.
			strconv.Itoa(u.ID),
			u.Name,
			u.NameKeyHash,
			u.FinishYear,
			csvBool(u.Professor),
			csvBool(u.TA),
			csvBool(u.StudentLeadership),
			csvBool(u.AlumniBoard),
		})
	}
}

func csvBool(b bool) string {
	if b {
		return "1"
	}

	return "0"
}
