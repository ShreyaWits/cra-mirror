package config

import (
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("EXISTING_KEY", "value123")
	defer os.Unsetenv("EXISTING_KEY")

	// Should return the set value
	if v := GetEnv("EXISTING_KEY", "default"); v != "value123" {
		t.Errorf("expected 'value123', got '%s'", v)
	}

	// Should return default for missing key
	if v := GetEnv("MISSING_KEY", "default"); v != "default" {
		t.Errorf("expected 'default', got '%s'", v)
	}
}

func TestLoadEnv(t *testing.T) {
	// Save original env
	orig := os.Environ()
	// Remove .env if exists, create a temp one
	f, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatalf("failed to create temp .env: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString("FOO=bar\n")
	f.Close()

	// Change working directory to temp dir containing .env
	oldWd, _ := os.Getwd()
	os.Chdir(os.TempDir())
	defer os.Chdir(oldWd)

	// Rename temp file to .env
	os.Rename(f.Name(), ".env")
	defer os.Remove(".env")

	// Capture log output
	old := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	LoadEnv()

	w.Close()
	log.SetOutput(old)
	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), "✅ Environment variables loaded successfully") {
		t.Errorf("expected success log, got: %s", string(out))
	}

	// Should set env var from .env
	if os.Getenv("FOO") != "bar" {
		t.Errorf("expected FOO=bar, got: %s", os.Getenv("FOO"))
	}

	// Restore env
	for _, kv := range orig {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			os.Setenv(parts[0], parts[1])
		}
	}
}

func TestLoadEnv_FileNotFound(t *testing.T) {
	// Remove .env if exists
	os.Remove(".env")

	// Capture log output
	old := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	LoadEnv()

	w.Close()
	log.SetOutput(old)
	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), "Error loading .env file") {
		t.Errorf("expected error log, got: %s", string(out))
	}
}
