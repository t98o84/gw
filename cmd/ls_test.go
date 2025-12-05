package cmd

import (
	"testing"
)

func TestLsCmd(t *testing.T) {
	if lsCmd == nil {
		t.Fatal("lsCmd should not be nil")
	}

	if lsCmd.Use != "ls" {
		t.Errorf("lsCmd.Use = %q, want %q", lsCmd.Use, "ls")
	}
}

func TestLsCmd_Aliases(t *testing.T) {
	aliases := lsCmd.Aliases
	expected := []string{"l"}

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

func TestLsCmd_PathFlag(t *testing.T) {
	flag := lsCmd.Flags().Lookup("path")
	if flag == nil {
		t.Fatal("Expected 'path' flag to be defined")
	}

	if flag.Shorthand != "p" {
		t.Errorf("path flag shorthand = %q, want %q", flag.Shorthand, "p")
	}
}
