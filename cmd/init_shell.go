package cmd

import (
	"embed"
	"fmt"
	"strings"

	"github.com/alde/haunt/internal/config"
	"github.com/spf13/cobra"
)

//go:embed shell/*
var shellScripts embed.FS

var initCmd = &cobra.Command{
	Use:       "init [fish|zsh]",
	Short:     "Print shell integration script",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"fish", "zsh"},
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		filename := fmt.Sprintf("shell/haunt.%s", shell)
		data, err := shellScripts.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("unsupported shell: %s (supported: fish, zsh)", shell)
		}

		binding, err := convertKeybinding(cfg.Keybinding, shell)
		if err != nil {
			return err
		}

		script := strings.ReplaceAll(string(data), "{{KEYBINDING}}", binding)
		fmt.Print(script)
		return nil
	},
}

func convertKeybinding(binding, shell string) (string, error) {
	parts := strings.Split(strings.ToLower(binding), "-")

	switch shell {
	case "fish":
		return fishBinding(parts)
	case "zsh":
		return zshBinding(parts)
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

func fishBinding(parts []string) (string, error) {
	if len(parts) == 2 && parts[0] == "ctrl" {
		return fmt.Sprintf("\\c%s", parts[1]), nil
	}
	if len(parts) == 2 && parts[0] == "alt" {
		return fmt.Sprintf("\\e%s", parts[1]), nil
	}
	if len(parts) == 3 && parts[0] == "ctrl" && parts[1] == "alt" {
		return fmt.Sprintf("\\e\\c%s", parts[2]), nil
	}
	return "", fmt.Errorf("unsupported keybinding format: %s", strings.Join(parts, "-"))
}

func zshBinding(parts []string) (string, error) {
	if len(parts) == 2 && parts[0] == "ctrl" {
		return fmt.Sprintf("^%s", parts[1]), nil
	}
	if len(parts) == 2 && parts[0] == "alt" {
		return fmt.Sprintf("^[%s", parts[1]), nil
	}
	if len(parts) == 3 && parts[0] == "ctrl" && parts[1] == "alt" {
		return fmt.Sprintf("^[^%s", parts[2]), nil
	}
	return "", fmt.Errorf("unsupported keybinding format: %s", strings.Join(parts, "-"))
}

func init() {
	rootCmd.AddCommand(initCmd)
}
