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

	// Delete removes the user info after it's been decrypted and used. It should be called only
	// after the successful use of data returned by Decrypt.
	Delete(id int) error
}

// ErrNotFound is returned by a Decrypter if the key did not decrypt any info.
var ErrNotFound = errors.New("user info not found")

// TableDecrypter implements Decrypter by using the key to attempt to decrypt all the users in the
// database.
type TableDecrypter struct {
	Table *db.UserTable
}

func (d TableDecrypter) Decrypt(key string) (*db.User, error) {
	keyHash, err := encrypt.MD5Hash(key)
	if err != nil {
		return nil, encrypt.NewErrBadKey(err)
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

		return user, nil
	}

	return nil, ErrNotFound
}

func (d TableDecrypter) Delete(id int) error {
	return d.Table.DeleteUser(context.Background(), id)
}
