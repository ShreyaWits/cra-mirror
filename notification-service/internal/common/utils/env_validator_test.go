package utils

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestValidateEnvVars_AllSet(t *testing.T) {
	os.Setenv("FOO", "bar")
	os.Setenv("BAZ", "qux")
	t.Cleanup(func() {
		os.Unsetenv("FOO")
		os.Unsetenv("BAZ")
	})

	// This should not exit or panic
	ValidateEnvVars([]string{"FOO", "BAZ"})
}

func TestValidateEnvVars_MissingVar(t *testing.T) {
	// Unset to ensure missing
	os.Unsetenv("SHOULD_NOT_EXIST")

	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess_MissingEnvVar")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	output, err := cmd.CombinedOutput()
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 1 {
			if !strings.Contains(string(output), "Required environment variable SHOULD_NOT_EXIST is not set") {
				t.Errorf("expected log output for missing env var, got: %s", string(output))
			}
			return
		}
	}
	t.Fatalf("process did not exit with code 1, err: %v, output: %s", err, string(output))
}

// TestHelperProcess_MissingEnvVar is not a real test. It's used as a subprocess for TestValidateEnvVars_MissingVar.
func TestHelperProcess_MissingEnvVar(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	ValidateEnvVars([]string{"SHOULD_NOT_EXIST"})
}
