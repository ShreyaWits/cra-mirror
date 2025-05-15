package cassandra

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocql/gocql"
)

func ApplyMigrations(session *gocql.Session, migrationsDir string) error {
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("could not read migrations directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".cql" {
			continue
		}

		version := strings.TrimSuffix(file.Name(), ".cql")
		var exists string
		err := session.Query("SELECT version FROM schema_migrations WHERE version = ?", version).Scan(&exists)
		if err == nil {
			log.Printf("✅ Migration %s already applied, skipping", version)
			continue
		}

		cqlBytes, err := ioutil.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return fmt.Errorf("error reading migration %s: %w", file.Name(), err)
		}

		queries := strings.Split(string(cqlBytes), ";")
		for _, q := range queries {
			q = strings.TrimSpace(q)
			if q == "" {
				continue
			}
			if err := session.Query(q).Exec(); err != nil {
				return fmt.Errorf("failed to execute query in %s: %w", file.Name(), err)
			}
		}

		if err := session.Query("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
			version, time.Now()).Exec(); err != nil {
			return fmt.Errorf("could not log migration %s: %w", version, err)
		}

		log.Printf("📦 Applied migration %s", version)
	}

	return nil
}
