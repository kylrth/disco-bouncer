package encrypt

import (
	//nolint:gosec // We use MD5 because we aren't concerned about collision attacks. We're only
	// using the hash for quickly searching for data that can be decrypted with this key.
	"crypto/md5"
	"encoding/hex"
)

// MD5Hash returns the MD5 hash of the provided key.
func MD5Hash(key string) (string, error) {
	bkey, err := hex.DecodeString(key)
	if err != nil {
		return "", err
	}

	bhash := md5.New().Sum(bkey) //nolint:gosec // See above.

	return hex.EncodeToString(bhash), nil
}
