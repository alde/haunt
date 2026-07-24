package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "haunt",
	Short: "Directory-aware shell history",
	Long:  "Haunt remembers what commands you ran and where. Your commands haunt the directories they were run in.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
