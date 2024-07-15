package main

import (
	"fmt"
	"os"

	"github.com/kylrth/disco-bouncer/pkg/encrypt"
	"github.com/spf13/cobra"
)

var runhashCmd = &cobra.Command{
	Use:   "runhash [KEYS...]",
	Short: "Compute the MD5 hash of the key exactly as is done on the server",
	Args:  cobra.ArbitraryArgs,
	Run: func(_ *cobra.Command, args []string) {
		err := runhash(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runhash(keys []string) error {
	for _, key := range keys {
		hash, err := encrypt.MD5Hash(key)
		if err != nil {
			return fmt.Errorf("hash key '%s': %w", key, err)
		}

		fmt.Println(hash)
	}

	return nil
}
