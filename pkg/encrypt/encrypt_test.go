package encrypt_test

import (
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kylrth/disco-bouncer/pkg/encrypt"
)

func ExampleEncrypt() {
	text := "some plain text"

	ciphertext, key, err := encrypt.Encrypt(text)
	if err != nil {
		panic(err)
	}
	out, err := encrypt.Decrypt(ciphertext, key)
	if err != nil {
		panic(err)
	}

	fmt.Println(out)
	// output:
	// some plain text
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	for i := range 20 {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			// get a random string of length between 8 and 1024
			length := rand.IntN(1024-8) + 8 //nolint:gosec // just a test
			b := make([]byte, length)
			_, err := crand.Read(b)
			if err != nil {
				t.Fatal(err)
			}

			plaintext := string(b)

			roundTrip(t, plaintext)
		})
	}
}

func roundTrip(t *testing.T, plaintext string) {
	t.Helper()

	ciphertext, key, err := encrypt.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	out, err := encrypt.Decrypt(ciphertext, key)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(plaintext, out); diff != "" {
		t.Error("unexpected output (-want +got):\n" + diff)
	}
}

func TestEncryptEdges(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"empty":     "",
		"non-latin": "민",
		"long":      strings.Repeat(" ", 2^15),
	}

	for name, s := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			roundTrip(t, s)
		})
	}
}

func TestDecrypt(t *testing.T) {
	t.Parallel()

	// generate an actual ciphertext and key
	plaintext := "some plain text here"
	ciphertext, key, err := encrypt.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	altPlain := "some other plain text; not the same"
	altCipher, altKey, err := encrypt.Encrypt(altPlain)
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		in  string
		err string
	}

	badKeyTests := map[string]testCase{
		"Nonhex":        {"asdfjklasioefniadsf", "encoding/hex: invalid byte: U+0073 's'"},
		"Size":          {"ab1f818344faecc111", "crypto/aes: invalid key size 9"},
		"WrongValidKey": {altKey, "cipher: message authentication failed"},
	}
	for name, c := range badKeyTests {
		t.Run("badKey"+name, func(t *testing.T) {
			t.Parallel()

			_, err := encrypt.Decrypt(ciphertext, c.in)

			if diff := cmp.Diff(c.err, err.Error()); diff != "" {
				t.Error("unexpected error (-want +got):\n" + diff)
			}
		})
	}

	badCipherTests := map[string]testCase{
		"Nonhex":           {"atdfifoewfijaba", "encoding/hex: invalid byte: U+0074 't'"},
		"Empty":            {"", "ciphertext too short"},
		"WrongValidCipher": {altCipher, "cipher: message authentication failed"},
	}
	for name, c := range badCipherTests {
		t.Run("badCipher"+name, func(t *testing.T) {
			t.Parallel()

			_, err := encrypt.Decrypt(c.in, key)

			if diff := cmp.Diff(c.err, err.Error()); diff != "" {
				t.Error("unexpected error (-want +got):\n" + diff)
			}
		})
	}
}

// Values copied from JavaScript output.
const (
	nonceLength = 12
	plaintext   = "Hello, World!"
	ciphertext  = "7486f7b175afecaa48f7e36c8148cfa748ce5d728f1320d8a5d2e5a9a06fa77e94ffd1d0cae76cb583"
	key         = "9fa17471ebcf4183d9fb76cde245acb09250bb5c31d2e8f51d5e8a6eb951eb1a"
)

func TestMatchJS(t *testing.T) {
	t.Parallel()

	bkey, err := hex.DecodeString(key)
	if err != nil {
		t.Fatal(err)
	}
	bciphertext, err := hex.DecodeString(ciphertext)
	if err != nil {
		t.Fatal(err)
	}

	out, err := encrypt.WithKeyAndNonce(plaintext, bkey, bciphertext[:nonceLength])
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(ciphertext, out); diff != "" {
		t.Error("unexpected ciphertext (-want +got):\n" + diff)
	}
}

func TestFromJS(t *testing.T) {
	t.Parallel()

	out, err := encrypt.Decrypt(ciphertext, key)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(plaintext, out); diff != "" {
		t.Error("unexpected output (-want +got):\n" + diff)
	}
}

func TestDecrypt_Errors(t *testing.T) {
	t.Parallel()

	_, err := encrypt.Decrypt("asdfjkl", key)
	if !errors.As(err, &encrypt.ErrBadCiphertext{}) {
		t.Errorf("error did not match ErrBadCiphertext: %v", err)
	}

	_, err = encrypt.Decrypt(ciphertext, "asdfjkl")
	if !errors.As(err, &encrypt.ErrBadKey{}) {
		t.Errorf("error did not match ErrBadKey: %v", err)
	}
}
