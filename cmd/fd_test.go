package cmd

import (
	"testing"
)

func TestFdCmd(t *testing.T) {
	if fdCmd == nil {
		t.Fatal("fdCmd should not be nil")
	}

	if fdCmd.Use != "fd" {
		t.Errorf("fdCmd.Use = %q, want %q", fdCmd.Use, "fd")
	}
}

func TestFdCmd_Aliases(t *testing.T) {
	aliases := fdCmd.Aliases
	expected := []string{"f"}

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

func TestFdCmd_PathFlag(t *testing.T) {
	flag := fdCmd.Flags().Lookup("path")
	if flag == nil {
		t.Fatal("Expected 'path' flag to be defined")
	}

	if flag.Shorthand != "p" {
		t.Errorf("path flag shorthand = %q, want %q", flag.Shorthand, "p")
	}
}
