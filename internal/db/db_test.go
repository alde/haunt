package db

import (
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestRecordAndSearchExact(t *testing.T) {
	store := openTestStore(t)

	store.Record("go build ./...", "/project", 0, 1000)
	store.Record("go test ./...", "/project", 0, 1001)
	store.Record("make deploy", "/other", 0, 1002)

	results, err := store.SearchExact("/project", 100)
	if err != nil {
		t.Fatalf("search: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d: %v", len(results), results)
	}

	if results[0] != "go test ./..." {
		t.Errorf("expected most recent first, got %q", results[0])
	}
	if results[1] != "go build ./..." {
		t.Errorf("expected second result %q, got %q", "go build ./...", results[1])
	}
}

func TestSearchExactDeduplicates(t *testing.T) {
	store := openTestStore(t)

	store.Record("make", "/project", 0, 1000)
	store.Record("make", "/project", 0, 1001)
	store.Record("make", "/project", 1, 1002)

	results, err := store.SearchExact("/project", 100)
	if err != nil {
		t.Fatalf("search: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 deduplicated result, got %d: %v", len(results), results)
	}
}

func TestSearchExactRespectsLimit(t *testing.T) {
	store := openTestStore(t)

	store.Record("cmd1", "/project", 0, 1000)
	store.Record("cmd2", "/project", 0, 1001)
	store.Record("cmd3", "/project", 0, 1002)

	results, err := store.SearchExact("/project", 2)
	if err != nil {
		t.Fatalf("search: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSearchExactNoResults(t *testing.T) {
	store := openTestStore(t)

	results, err := store.SearchExact("/nonexistent", 100)
	if err != nil {
		t.Fatalf("search: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearchGitRoot(t *testing.T) {
	store := openTestStore(t)

	store.Record("go build", "/repo", 0, 1000)
	store.Record("go test", "/repo/pkg", 0, 1001)
	store.Record("go vet", "/repo/pkg/sub", 0, 1002)
	store.Record("unrelated", "/other", 0, 1003)

	results, err := store.search(func(e *Entry) bool {
		return e.Cwd == "/repo" || len(e.Cwd) > len("/repo") && e.Cwd[:len("/repo")+1] == "/repo/"
	}, 100)
	if err != nil {
		t.Fatalf("search: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d: %v", len(results), results)
	}
}

func TestAncestorDirs(t *testing.T) {
	tests := []struct {
		name    string
		cwd     string
		ceiling string
		want    []string
	}{
		{
			name:    "walks up to ceiling",
			cwd:     "/home/user/project/src/pkg",
			ceiling: "/home/user/project",
			want:    []string{"/home/user/project/src/pkg", "/home/user/project/src", "/home/user/project"},
		},
		{
			name:    "cwd is ceiling",
			cwd:     "/home/user/project",
			ceiling: "/home/user/project",
			want:    []string{"/home/user/project"},
		},
		{
			name:    "root ceiling",
			cwd:     "/home",
			ceiling: "/",
			want:    []string{"/home", "/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ancestorDirsWithCeiling(tt.cwd, tt.ceiling)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSearchAncestors(t *testing.T) {
	store := openTestStore(t)

	store.Record("repo-root-cmd", "/project", 0, 1000)
	store.Record("src-cmd", "/project/src", 0, 1001)
	store.Record("deep-cmd", "/project/src/pkg/api", 0, 1002)
	store.Record("unrelated", "/other", 0, 1003)

	dirs := make(map[string]bool)
	for _, d := range ancestorDirsWithCeiling("/project/src/pkg/api", "/project") {
		dirs[d] = true
	}
	results, err := store.search(func(e *Entry) bool {
		return dirs[e.Cwd]
	}, 100)
	if err != nil {
		t.Fatalf("search: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d: %v", len(results), results)
	}

	if results[0] != "deep-cmd" {
		t.Errorf("expected most recent first, got %q", results[0])
	}
}
