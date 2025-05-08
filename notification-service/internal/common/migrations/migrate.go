package migrations

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gocql/gocql"
)

type Queryable interface {
	Exec() error
}

type QueryableSession interface {
	Query(string, ...interface{}) Queryable
}

// Adapter for *gocql.Query to Queryable
type gocqlQueryAdapter struct {
	q *gocql.Query
}

func (a *gocqlQueryAdapter) Exec() error { return a.q.Exec() }

// Adapter for *gocql.Session to QueryableSession
type gocqlSessionAdapter struct {
	s *gocql.Session
}

func (a *gocqlSessionAdapter) Query(stmt string, values ...interface{}) Queryable {
	return &gocqlQueryAdapter{q: a.s.Query(stmt, values...)}
}

// Refactored for testability
func RunMigrationsWithDeps(
	session QueryableSession,
	dir string,
	readDir func(string) ([]os.DirEntry, error),
	readFile func(string) ([]byte, error),
	fatalf func(string, ...interface{}),
) {
	files, err := readDir(dir)
	if err != nil {
		fatalf("Could not read migrations directory: %v", err)
	}

	log.Println("Found migration files", files)

	// Sort files to ensure migrations run in order
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".cql") {
			content, err := readFile(filepath.Join(dir, file.Name()))
			if err != nil {
				fatalf("Could not read migration file %s: %v", file.Name(), err)
			}

			if err := session.Query(string(content)).Exec(); err != nil {
				fatalf("Migration %s failed: %v", file.Name(), err)
			}

			log.Printf("Successfully ran migration: %s", file.Name())
		}
	}
}

// Default for production use
func RunMigrations(session *gocql.Session) {
	RunMigrationsWithDeps(&gocqlSessionAdapter{s: session}, "internal/common/migrations", os.ReadDir, os.ReadFile, log.Fatalf)
}
