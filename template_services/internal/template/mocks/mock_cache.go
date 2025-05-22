package mocks

import (
	"time"
	"template-services/pkg/cache"
)

// MockCache is a mock implementation of CacheInterface
type MockCache struct {
	GetFunc    func(key string, value interface{}) error
	SetFunc    func(key string, value interface{}, ttl time.Duration) error
	DeleteFunc func(key string) error
}

// Ensure MockCache implements CacheInterface
var _ cache.CacheInterface = (*MockCache)(nil)

func (m *MockCache) Get(key string, value interface{}) error {
	if m.GetFunc != nil {
		return m.GetFunc(key, value)
	}
	return nil
}

func (m *MockCache) Set(key string, value interface{}, ttl time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, value, ttl)
	}
	return nil
}

func (m *MockCache) Delete(key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(key)
	}
	return nil
} 
