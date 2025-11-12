// Package encrypt implements AES encryption helper functions for use with the bouncerbot.
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

const nonceLength = 12

// Encrypt encodes plain text into a ciphertext using a randomly-generated key (which is then
// returned as a hexadecimal string).
func Encrypt(text string) (ciphertext, key string, err error) {
	bkey, err := generateKey()
	if err != nil {
		return "", "", fmt.Errorf("generate key: %w", err)
	}
	key = hex.EncodeToString(bkey)

	nonce := make([]byte, nonceLength)
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", key, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext, err = WithKeyAndNonce(text, bkey, nonce)

	return ciphertext, key, err
}

func generateKey() ([]byte, error) {
	key := make([]byte, 32) // 32 bytes for 256-bit key
	_, err := rand.Read(key)

	return key, err
}

// WithKeyAndNonce encrypts the text using the same algorithm as Encrypt, but with the given key and
// nonce. It is the caller's responsibility to ensure the key and nonce are securely generated and
// shared.
func WithKeyAndNonce(text string, key, nonce []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	bciphertext := aesgcm.Seal(nil, nonce, []byte(text), nil)
	ciphertext := hex.EncodeToString(nonce) + hex.EncodeToString(bciphertext)

	return ciphertext, nil
}

// Decrypt decodes plain text from the ciphertext using the provided key. Non-nil errors are either
// BadCiphertextError, BadKeyError, or InauthenticatedError.
func Decrypt(ciphertext, key string) (string, error) {
	bciphertext, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", BadCiphertextError{err}
	}
	if len(bciphertext) < nonceLength {
		return "", BadCiphertextError{errors.New("ciphertext too short")}
	}

	bkey, err := hex.DecodeString(key)
	if err != nil {
		return "", BadKeyError{err}
	}
	block, err := aes.NewCipher(bkey)
	if err != nil {
		return "", BadKeyError{err}
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", BadKeyError{err}
	}

	btext, err := aesgcm.Open(nil, bciphertext[:nonceLength], bciphertext[nonceLength:], nil)
	if err != nil {
		return string(btext), InauthenticatedError{err}
	}

	return string(btext), nil
}

// BadCiphertextError is returned if the ciphertext is invalid.
type BadCiphertextError struct {
	error
}

// BadKeyError is returned if the key is invalid.
type BadKeyError struct {
	error
}

// NewBadKeyError wraps the given error to signify that this error was caused by a bad key.
func NewBadKeyError(wrap error) BadKeyError {
	return BadKeyError{wrap}
}

// InauthenticatedError is returned when the key was not the correct one for the ciphertext.
type InauthenticatedError struct {
	error
}
