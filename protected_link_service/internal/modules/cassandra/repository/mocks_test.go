package repository

import (
	cassandra "protected_link/pkg/cassandra"

	"github.com/gocql/gocql"
)

// --- MockQuery for cassandra.QueryInterface ---
type MockQuery struct {
	ScanFunc func(...interface{}) error
	ExecFunc func() error
}

func (m *MockQuery) Consistency(_ gocql.Consistency) cassandra.QueryInterface {
	return m
}

func (m *MockQuery) Scan(dest ...interface{}) error {
	if m.ScanFunc != nil {
		return m.ScanFunc(dest...)
	}
	return nil
}

func (m *MockQuery) Exec() error {
	if m.ExecFunc != nil {
		return m.ExecFunc()
	}
	return nil
}

// --- MockSession for cassandra.SessionInterface ---
type MockSession struct {
	QueryFunc func(string, ...interface{}) cassandra.QueryInterface
	CloseFunc func()
}

func (m *MockSession) Query(stmt string, args ...interface{}) cassandra.QueryInterface {
	return m.QueryFunc(stmt, args...)
}

func (m *MockSession) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
