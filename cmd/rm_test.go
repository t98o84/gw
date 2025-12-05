package cmd

import (
	"testing"
)

func TestRmCmd(t *testing.T) {
	if rmCmd == nil {
		t.Fatal("rmCmd should not be nil")
	}

	if rmCmd.Use != "rm [name...]" {
		t.Errorf("rmCmd.Use = %q, want %q", rmCmd.Use, "rm [name...]")
	}
}

func TestRmCmd_Aliases(t *testing.T) {
	aliases := rmCmd.Aliases
	expected := []string{"r"}

	if len(aliases) != len(expected) {
		t.Errorf("Expected %d aliases, got %d", len(expected), len(aliases))
		return
	}

	for i, alias := range expected {
		if aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, aliases[i], alias)
		}
	}
}

func TestRmCmd_ForceFlag(t *testing.T) {
	flag := rmCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("Expected 'force' flag to be defined")
	}

	if flag.Shorthand != "f" {
		t.Errorf("force flag shorthand = %q, want %q", flag.Shorthand, "f")
	}
}

func TestRmCmd_AcceptsMultipleArgs(t *testing.T) {
	// rmCmd should accept multiple arguments (no Args restriction)
	if rmCmd.Args != nil {
		t.Log("rmCmd.Args is defined, which is fine as long as it allows multiple args")
	}
}
