package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Scope string

const (
	ScopeExact     Scope = "exact"
	ScopeAncestors Scope = "ancestors"
	ScopeGitRoot   Scope = "git-root"
)

type Config struct {
	Keybinding string `toml:"keybinding"`
	Scope      Scope  `toml:"scope"`
	DBPath     string `toml:"db_path"`
}

func Default() Config {
	return Config{
		Keybinding: "ctrl-g",
		Scope:      ScopeAncestors,
		DBPath:     filepath.Join(dataDir(), "haunt", "history.db"),
	}
}

func Load() (Config, error) {
	cfg := Default()

	path := filepath.Join(configDir(), "haunt", "config.toml")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	defer f.Close()

	if _, err := toml.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, err
	}

	cfg.DBPath = expandHome(cfg.DBPath)
	return cfg, nil
}

func configDir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}

func dataDir() string {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share")
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
