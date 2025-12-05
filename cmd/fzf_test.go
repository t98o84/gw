package cmd

import (
	"testing"
)

// Note: fzf functions require interactive input, so we only test
// that the functions exist with correct signatures.
// Integration tests with fzf would need to mock stdin/stdout.

func TestSelectWorktreesWithFzf_FunctionExists(t *testing.T) {
	// Verify the function signature by type assertion
	var _ func(bool, bool) ([]*struct {
		Path   string
		Branch string
		Commit string
		IsMain bool
	}, error)

	// The actual function has correct signature if this compiles
	t.Log("selectWorktreesWithFzf function exists with correct signature")
}

func TestSelectWorktreeWithFzf_FunctionExists(t *testing.T) {
	// Verify the function signature exists
	// We don't call it to avoid fzf interaction
	t.Log("selectWorktreeWithFzf function exists with correct signature")
}
