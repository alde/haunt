package cmd

import (
	"fmt"

	"github.com/alde/haunt/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		fmt.Printf("keybinding = %q\n", cfg.Keybinding)
		fmt.Printf("scope      = %q\n", cfg.Scope)
		fmt.Printf("db_path    = %q\n", cfg.DBPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
