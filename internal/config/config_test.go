package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Keybinding != "ctrl-g" {
		t.Errorf("keybinding = %q, want %q", cfg.Keybinding, "ctrl-g")
	}
	if cfg.Scope != ScopeAncestors {
		t.Errorf("scope = %q, want %q", cfg.Scope, ScopeAncestors)
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Keybinding != "ctrl-g" {
		t.Errorf("expected defaults when config missing, got keybinding=%q", cfg.Keybinding)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", dir)

	configDir := filepath.Join(dir, "haunt")
	os.MkdirAll(configDir, 0o755)

	content := `
keybinding = "alt-r"
scope = "git-root"
db_path = "~/custom/history.db"
`
	os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(content), 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Keybinding != "alt-r" {
		t.Errorf("keybinding = %q, want %q", cfg.Keybinding, "alt-r")
	}
	if cfg.Scope != ScopeGitRoot {
		t.Errorf("scope = %q, want %q", cfg.Scope, ScopeGitRoot)
	}

	home, _ := os.UserHomeDir()
	want := filepath.Join(home, "custom/history.db")
	if cfg.DBPath != want {
		t.Errorf("db_path = %q, want %q", cfg.DBPath, want)
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/foo/bar", filepath.Join(home, "foo/bar")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		got := expandHome(tt.input)
		if got != tt.want {
			t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
