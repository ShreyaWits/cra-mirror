package migrations

import (
	"errors"
	"os"
	"strings"
	"testing"
)

type mockDirEntry struct{ name string }

func (m mockDirEntry) Name() string               { return m.name }
func (m mockDirEntry) IsDir() bool                { return false }
func (m mockDirEntry) Type() os.FileMode          { return 0 }
func (m mockDirEntry) Info() (os.FileInfo, error) { return nil, nil }

type mockQuery struct {
	fail bool
}

func (q *mockQuery) Exec() error {
	if q.fail {
		return errors.New("query failed")
	}
	return nil
}

type mockSession struct {
	failQuery bool
}

func (m *mockSession) Query(stmt string, values ...interface{}) Queryable {
	return &mockQuery{fail: m.failQuery}
}

func TestRunMigrationsWithDeps(t *testing.T) {
	dir := "test_migrations"
	files := []os.DirEntry{
		mockDirEntry{"001_test.cql"},
		mockDirEntry{"002_test.cql"},
		mockDirEntry{"not_a_migration.txt"},
	}

	t.Run("directory read error", func(t *testing.T) {
		fatalCalled := false
		defer func() { recover() }()
		RunMigrationsWithDeps(
			&mockSession{},
			dir,
			func(string) ([]os.DirEntry, error) { return nil, errors.New("fail dir") },
			os.ReadFile,
			func(string, ...interface{}) { fatalCalled = true; panic("fatal") },
		)
		if !fatalCalled {
			t.Error("expected fatal for directory read error")
		}
	})

	t.Run("file read error", func(t *testing.T) {
		fatalCalled := false
		defer func() { recover() }()
		RunMigrationsWithDeps(
			&mockSession{},
			dir,
			func(string) ([]os.DirEntry, error) { return files, nil },
			func(path string) ([]byte, error) {
				if strings.Contains(path, "001_test.cql") {
					return nil, errors.New("fail file")
				}
				return []byte("ok"), nil
			},
			func(string, ...interface{}) { fatalCalled = true; panic("fatal") },
		)
		if !fatalCalled {
			t.Error("expected fatal for file read error")
		}
	})

	t.Run("query error", func(t *testing.T) {
		fatalCalled := false
		defer func() { recover() }()
		RunMigrationsWithDeps(
			&mockSession{failQuery: true},
			dir,
			func(string) ([]os.DirEntry, error) { return files, nil },
			func(path string) ([]byte, error) { return []byte("SELECT 1;"), nil },
			func(string, ...interface{}) { fatalCalled = true; panic("fatal") },
		)
		if !fatalCalled {
			t.Error("expected fatal for query error")
		}
	})

	t.Run("success", func(t *testing.T) {
		fatalCalled := false
		RunMigrationsWithDeps(
			&mockSession{failQuery: false},
			dir,
			func(string) ([]os.DirEntry, error) { return files, nil },
			func(path string) ([]byte, error) { return []byte("SELECT 1;"), nil },
			func(string, ...interface{}) { fatalCalled = true; panic("fatal") },
		)
		if fatalCalled {
			t.Error("did not expect fatal for success")
		}
	})
}
