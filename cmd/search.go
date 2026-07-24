package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alde/haunt/internal/config"
	"github.com/alde/haunt/internal/db"
	"github.com/spf13/cobra"
)

const defaultLimit = 5000

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search directory-aware history",
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

		results, err := search(store, cfg.Scope, cwd)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			return nil
		}

		selected, err := fzf(results)
		if err != nil {
			return err
		}

		fmt.Print(selected)
		return nil
	},
}

func search(store *db.Store, scope config.Scope, cwd string) ([]string, error) {
	switch scope {
	case config.ScopeExact:
		return store.SearchExact(cwd, defaultLimit)
	case config.ScopeGitRoot:
		return store.SearchGitRoot(cwd, defaultLimit)
	default:
		return store.SearchAncestors(cwd, defaultLimit)
	}
}

func fzf(items []string) (string, error) {
	cmd := exec.Command("fzf", "--no-sort", "--exact", "--prompt", "haunt> ")
	cmd.Stdin = strings.NewReader(strings.Join(items, "\n"))
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && (exitErr.ExitCode() == 1 || exitErr.ExitCode() == 130) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
