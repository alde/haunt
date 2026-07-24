package db

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type Entry struct {
	Command   string
	Cwd       string
	Timestamp int64
	ExitCode  int
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(wal)")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrating database: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			command   TEXT    NOT NULL,
			cwd       TEXT    NOT NULL,
			timestamp INTEGER NOT NULL,
			exit_code INTEGER NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_history_cwd ON history(cwd);
		CREATE INDEX IF NOT EXISTS idx_history_timestamp ON history(timestamp DESC);
	`)
	return err
}

func (s *Store) Record(command, cwd string, exitCode int, timestamp int64) error {
	_, err := s.db.Exec(
		`INSERT INTO history (command, cwd, timestamp, exit_code) VALUES (?, ?, ?, ?)`,
		command, cwd, timestamp, exitCode,
	)
	return err
}

func (s *Store) SearchExact(cwd string, limit int) ([]string, error) {
	return s.queryDistinct(`WHERE cwd = ?`, limit, cwd)
}

func (s *Store) SearchAncestors(cwd string, limit int) ([]string, error) {
	return s.searchDirs(ancestorDirs(cwd), limit)
}

func (s *Store) searchDirs(dirs []string, limit int) ([]string, error) {
	placeholders := make([]string, len(dirs))
	args := make([]any, len(dirs))
	for i, d := range dirs {
		placeholders[i] = "?"
		args[i] = d
	}

	where := fmt.Sprintf("WHERE cwd IN (%s)", strings.Join(placeholders, ","))
	return s.queryDistinct(where, limit, args...)
}

func (s *Store) SearchGitRoot(cwd string, limit int) ([]string, error) {
	root := gitRoot(cwd)
	return s.queryDistinct(`WHERE cwd = ? OR cwd LIKE ?`, limit, root, root+"/%")
}

func (s *Store) queryDistinct(where string, limit int, args ...any) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT command FROM history %s
		GROUP BY command
		ORDER BY MAX(timestamp) DESC
		LIMIT ?
	`, where)
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var cmd string
		if err := rows.Scan(&cmd); err != nil {
			return nil, err
		}
		results = append(results, cmd)
	}
	return results, rows.Err()
}

func ancestorDirs(cwd string) []string {
	return ancestorDirsWithCeiling(cwd, gitRoot(cwd))
}

func ancestorDirsWithCeiling(cwd, ceiling string) []string {
	var dirs []string
	dir := cwd
	for {
		dirs = append(dirs, dir)
		if dir == ceiling || dir == "/" || dir == "." {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return dirs
}

func gitRoot(cwd string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return cwd
	}
	return strings.TrimSpace(string(out))
}
