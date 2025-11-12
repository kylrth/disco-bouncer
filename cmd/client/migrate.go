package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cobaltspeech/log"
	"github.com/kylrth/disco-bouncer/pkg/client"
	"github.com/kylrth/disco-bouncer/pkg/encrypt"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate ROLE",
	Short: "Bulk migrate students to a new cohort",
	Long: `Move a list of students from the "pre-core ACME" role to the specified cohort year.

The year can include non-digit characters as in "2026w", but it must match an existing role that
begins with that string.

The student info is accepted in CSV format on stdin. The first line may optionally be exactly the
header below:

	id,name,key
	12,John Doe,1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcd
	...

The key is *never* sent to the server, but is used to find and change the user's information on the
server. This command tries to update on the server first, and then if not present tries to update
an existing Discord user's roles. In that case, the name must exactly match the nickname of a user
on the server who currently has the "pre-core ACME" role.

The "id" field is discarded, and only accepted for compatibility with the output of the 'upload'
command. If a key is missing for a particular user, you can leave the key field empty for that row
and the name will be migrated on Discord.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if onlyNames && onlyKeys {
			return errors.New("cannot use both --only-names and --only-keys")
		}

		return cobra.ExactArgs(1)(cmd, args)
	},
	Run: withLAndC(func(l log.Logger, c *client.Client, s []string) error {
		if onlyNames || onlyKeys {
			return migrateOnly(l, c, s[0])
		}

		return migrate(l, c, s[0])
	}),
}

var (
	onlyNames bool
	onlyKeys  bool
)

func init() {
	migrateCmd.Flags().BoolVar(
		&onlyNames, "only-names", false, "stdin contains only names, one on each line (no "+
			"header). This will only migrate existing users on Discord.",
	)
	migrateCmd.Flags().BoolVar(
		&onlyKeys, "only-keys", false, "stdin contains only keys, one on each line (no header). "+
			"This will only migrate uploaded users who have not joined Discord yet.",
	)
}

func isMigrateHeader(line []string) bool {
	return line[0] == "id" &&
		line[1] == "name" &&
		line[2] == "key"
}

func migrate(l log.Logger, c *client.Client, year string) error {
	r := csv.NewReader(os.Stdin)
	r.ReuseRecord = true
	r.FieldsPerRecord = 3

	lineNum := 0
	for {
		line, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return err
		}
		lineNum++

		if lineNum == 1 && isMigrateHeader(line) {
			continue
		}

		err = migrateTryBoth(context.Background(), l, c, line[1], line[2], year)
		if err != nil {
			return err
		}
	}
}

func migrateTryBoth(
	ctx context.Context, l log.Logger, c *client.Client, name, key, year string,
) error {
	if key == "" {
		l.Debug("msg", "empty key; trying Discord", "name", name)
	} else {
		err := migrateByKey(ctx, l, c, key, year)
		if !errors.Is(err, ErrNotFound) {
			return err
		}
		l.Debug("msg", "user not found by key; trying Discord", "name", name, "key", key)
	}

	return c.Discord.Migrate(ctx, name, year)
}

func migrateOnly(l log.Logger, c *client.Client, year string) error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}

		ctx := context.Background()
		var with string // for logging
		var err error

		if onlyNames {
			with = "name " + text
			err = c.Discord.Migrate(ctx, text, year)
		} else {
			with = "key " + text
			err = migrateByKey(ctx, l, c, text, year)
		}

		if err != nil {
			return fmt.Errorf("failed to migrate user with "+with+": %w", err)
		}
	}

	return nil
}

func migrateByKey(ctx context.Context, l log.Logger, c *client.Client, key, year string) error {
	// Do the same as getWithKey but keep the name ciphertext the same. We don't want to reupload
	// the plaintext name when we update the FinishYear.
	hash, err := encrypt.MD5Hash(key)
	if err != nil {
		return err
	}

	users, err := c.Users.GetAllUsers(ctx, client.WithKeyHash(hash))
	if err != nil {
		return err
	}

	for _, u := range users {
		_, err = encrypt.Decrypt(u.Name, key)
		if errors.As(err, &encrypt.InauthenticatedError{}) {
			// wrong key (wow, hash collision!)
			continue
		}
		if err != nil {
			return err
		}

		l.Debug("msg", "found user by key hash", "key", key, "id", u.ID)

		u.FinishYear = year

		err = c.Users.UpdateUser(ctx, u)
		if err != nil {
			return err
		}

		l.Debug("msg", "updated user", "id", u.ID, "year", year)

		return nil
	}

	return ErrNotFound
}
