package repository

import "github.com/gocql/gocql"

type QueryInterface interface {
	Consistency(gocql.Consistency) QueryInterface
	Scan(...interface{}) error
	Exec() error
}

type SessionInterface interface {
	Query(string, ...interface{}) QueryInterface
	Close()
}
