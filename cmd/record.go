package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alde/haunt/internal/config"
	"github.com/alde/haunt/internal/db"
	"github.com/spf13/cobra"
)

var recordExitCode int

var recordCmd = &cobra.Command{
	Use:    "record [command...]",
	Short:  "Record a command to history",
	Hidden: true,
	Args:   cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		store, err := db.Open(cfg.DBPath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer store.Close()

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting cwd: %w", err)
		}

		command := strings.Join(args, " ")
		return store.Record(command, cwd, recordExitCode, time.Now().Unix())
	},
}

func init() {
	recordCmd.Flags().IntVar(&recordExitCode, "exit-code", 0, "exit code of the command")
	rootCmd.AddCommand(recordCmd)
}
