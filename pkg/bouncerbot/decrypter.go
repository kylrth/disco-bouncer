package bouncerbot

import (
	"context"
	"errors"
	"fmt"

	"github.com/kylrth/disco-bouncer/internal/db"
	"github.com/kylrth/disco-bouncer/pkg/encrypt"
)

// Decrypter provides user information conditioned on receiving the decryption key for it.
type Decrypter interface {
	// Decrypt attempts to decrypt any user info using the key provided. It returns ErrNotFound if
	// the key did not decrypt anything.
	Decrypt(key string) (*db.User, error)
}

// ErrNotFound is returned by a Decrypter if the key did not decrypt any info.
var ErrNotFound = errors.New("user info not found")

// TableDecrypter implements Decrypter by using the key to attempt to decrypt all the users in the
// database. If the user is found, it is deleted from the table.
type TableDecrypter struct {
	Table *db.UserTable
}

func (d TableDecrypter) Decrypt(key string) (*db.User, error) {
	keyHash, err := encrypt.MD5Hash(key)
	if err != nil {
		return nil, err
	}

	users, err := d.Table.GetUsers(context.Background(), db.WithKeyHash(keyHash))
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}

	for _, user := range users {
		user.Name, err = encrypt.Decrypt(user.Name, key)
		if err != nil {
			if errors.As(err, &encrypt.ErrBadKey{}) {
				return nil, err
			}

			// wrong key for this user
			continue
		}

		// Delete the user from the table.
		err = d.Table.DeleteUser(context.Background(), user.ID)
		if err != nil {
			return user, fmt.Errorf("delete user after decryption: %w", err)
		}

		return user, nil
	}

	return nil, ErrNotFound
}
