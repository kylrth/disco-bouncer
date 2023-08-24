package main

import (
	"fmt"
	"os"

	"github.com/cobaltspeech/log"
	"github.com/cobaltspeech/log/pkg/level"
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

var verbosity int

func init() {
	rootCmd.AddCommand(
		serveCmd,
		adminCmd,
	)

	rootCmd.PersistentFlags().IntVarP(&verbosity, "verbosity", "v", 2, "set verbosity (1-4)")
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
