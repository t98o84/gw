package shell

import (
	"os"
	"os/exec"
)

// Executor abstracts shell command execution
type Executor interface {
	// Execute runs a command and returns its standard output
	Execute(name string, args ...string) ([]byte, error)

	// ExecuteWithStdio runs a command with connected standard I/O
	ExecuteWithStdio(name string, args ...string) error

	// LookPath checks if a command exists
	LookPath(name string) (string, error)
}

// RealExecutor implements actual command execution
type RealExecutor struct{}

// NewRealExecutor creates a new RealExecutor
func NewRealExecutor() *RealExecutor {
	return &RealExecutor{}
}

// Execute runs a command and returns its output
func (e *RealExecutor) Execute(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

// ExecuteWithStdio runs a command with connected standard I/O
func (e *RealExecutor) ExecuteWithStdio(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// LookPath checks if a command exists in PATH
func (e *RealExecutor) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}
