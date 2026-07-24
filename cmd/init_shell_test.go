package cmd

import "testing"

func TestConvertKeybinding(t *testing.T) {
	tests := []struct {
		binding string
		shell   string
		want    string
	}{
		{"ctrl-g", "fish", `\cg`},
		{"ctrl-g", "zsh", "^g"},
		{"alt-r", "fish", `\er`},
		{"alt-r", "zsh", "^[r"},
		{"ctrl-alt-r", "fish", `\e\cr`},
		{"ctrl-alt-r", "zsh", "^[^r"},
		{"CTRL-G", "fish", `\cg`},
	}

	for _, tt := range tests {
		t.Run(tt.binding+"_"+tt.shell, func(t *testing.T) {
			got, err := convertKeybinding(tt.binding, tt.shell)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConvertKeybindingInvalid(t *testing.T) {
	_, err := convertKeybinding("invalid", "fish")
	if err == nil {
		t.Error("expected error for invalid binding")
	}
}
