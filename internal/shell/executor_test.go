package shell

import (
	"os/exec"
	"strings"
	"testing"
)

// TestRealExecutor_Execute tests command execution with output capture
func TestRealExecutor_Execute(t *testing.T) {
	executor := NewRealExecutor()

	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
		wantOut string
	}{
		{
			name:    "echo command success",
			command: "echo",
			args:    []string{"hello"},
			wantErr: false,
			wantOut: "hello",
		},
		{
			name:    "non-existent command",
			command: "nonexistentcommand12345",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "command with multiple args",
			command: "echo",
			args:    []string{"hello", "world"},
			wantErr: false,
			wantOut: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := executor.Execute(tt.command, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.wantOut != "" {
				got := strings.TrimSpace(string(out))
				if got != tt.wantOut {
					t.Errorf("Execute() = %v, want %v", got, tt.wantOut)
				}
			}
		})
	}
}

// TestRealExecutor_ExecuteWithStdio tests command execution with stdio connection
func TestRealExecutor_ExecuteWithStdio(t *testing.T) {
	executor := NewRealExecutor()

	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
	}{
		{
			name:    "true command success",
			command: "true",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "false command fails",
			command: "false",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "non-existent command",
			command: "nonexistentcommand12345",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.ExecuteWithStdio(tt.command, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteWithStdio() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRealExecutor_LookPath tests command existence checking
func TestRealExecutor_LookPath(t *testing.T) {
	executor := NewRealExecutor()

	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "echo exists",
			command: "echo",
			wantErr: false,
		},
		{
			name:    "sh exists",
			command: "sh",
			wantErr: false,
		},
		{
			name:    "non-existent command",
			command: "nonexistentcommand12345",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := executor.LookPath(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && path == "" {
				t.Errorf("LookPath() returned empty path for existing command %s", tt.command)
			}
		})
	}
}

// TestRealExecutor_Execute_ErrorHandling tests error handling
func TestRealExecutor_Execute_ErrorHandling(t *testing.T) {
	executor := NewRealExecutor()

	// Test command that returns error
	_, err := executor.Execute("sh", "-c", "exit 1")
	if err == nil {
		t.Error("Execute() expected error for command with non-zero exit code")
	}

	// Verify it's an ExitError
	if _, ok := err.(*exec.ExitError); !ok {
		t.Errorf("Execute() error type = %T, want *exec.ExitError", err)
	}
}

// TestMockExecutor verifies the mock works as expected
func TestMockExecutor(t *testing.T) {
	t.Run("default behavior", func(t *testing.T) {
		mock := &MockExecutor{}

		out, err := mock.Execute("test")
		if err != nil {
			t.Errorf("Execute() unexpected error: %v", err)
		}
		if string(out) != "mock output" {
			t.Errorf("Execute() = %s, want 'mock output'", out)
		}

		err = mock.ExecuteWithStdio("test")
		if err != nil {
			t.Errorf("ExecuteWithStdio() unexpected error: %v", err)
		}

		path, err := mock.LookPath("test")
		if err != nil {
			t.Errorf("LookPath() unexpected error: %v", err)
		}
		if path != "/usr/bin/test" {
			t.Errorf("LookPath() = %s, want '/usr/bin/test'", path)
		}
	})

	t.Run("custom behavior", func(t *testing.T) {
		mock := &MockExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("custom"), nil
			},
		}

		out, err := mock.Execute("test")
		if err != nil {
			t.Errorf("Execute() unexpected error: %v", err)
		}
		if string(out) != "custom" {
			t.Errorf("Execute() = %s, want 'custom'", out)
		}
	})
}
