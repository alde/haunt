package db

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	bolt "go.etcd.io/bbolt"
)

var historyBucket = []byte("history")

type Store struct {
	db *bolt.DB
}

type Entry struct {
	Command   string `json:"command"`
	Cwd       string `json:"cwd"`
	Timestamp int64  `json:"timestamp"`
	ExitCode  int    `json:"exit_code"`
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	bdb, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	err = bdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(historyBucket)
		return err
	})
	if err != nil {
		bdb.Close()
		return nil, fmt.Errorf("creating bucket: %w", err)
	}

	return &Store{db: bdb}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Record(command, cwd string, exitCode int, timestamp int64) error {
	entry := Entry{
		Command:   command,
		Cwd:       cwd,
		Timestamp: timestamp,
		ExitCode:  exitCode,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshalling entry: %w", err)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(historyBucket)
		seq, err := b.NextSequence()
		if err != nil {
			return err
		}
		return b.Put(uint64ToBytes(seq), data)
	})
}

func (s *Store) SearchExact(cwd string, limit int) ([]string, error) {
	return s.search(func(e *Entry) bool {
		return e.Cwd == cwd
	}, limit)
}

func (s *Store) SearchAncestors(cwd string, limit int) ([]string, error) {
	dirs := make(map[string]bool)
	for _, d := range ancestorDirs(cwd) {
		dirs[d] = true
	}
	return s.search(func(e *Entry) bool {
		return dirs[e.Cwd]
	}, limit)
}

func (s *Store) SearchGitRoot(cwd string, limit int) ([]string, error) {
	root := gitRoot(cwd)
	return s.search(func(e *Entry) bool {
		return e.Cwd == root || strings.HasPrefix(e.Cwd, root+"/")
	}, limit)
}

func (s *Store) search(match func(*Entry) bool, limit int) ([]string, error) {
	type cmdInfo struct {
		maxTimestamp int64
	}
	seen := make(map[string]*cmdInfo)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(historyBucket)
		return b.ForEach(func(_, v []byte) error {
			var e Entry
			if err := json.Unmarshal(v, &e); err != nil {
				return nil
			}
			if !match(&e) {
				return nil
			}
			if info, ok := seen[e.Command]; ok {
				if e.Timestamp > info.maxTimestamp {
					info.maxTimestamp = e.Timestamp
				}
			} else {
				seen[e.Command] = &cmdInfo{maxTimestamp: e.Timestamp}
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	type ranked struct {
		command      string
		maxTimestamp int64
	}
	items := make([]ranked, 0, len(seen))
	for cmd, info := range seen {
		items = append(items, ranked{command: cmd, maxTimestamp: info.maxTimestamp})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].maxTimestamp > items[j].maxTimestamp
	})

	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	results := make([]string, len(items))
	for i, item := range items {
		results[i] = item.command
	}
	return results, nil
}

func uint64ToBytes(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
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
