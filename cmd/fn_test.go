package cmd

import (
	"testing"
)

func TestFnCmd(t *testing.T) {
	if fnCmd == nil {
		t.Fatal("fnCmd should not be nil")
	}

	if fnCmd.Use != "fn" {
		t.Errorf("fnCmd.Use = %q, want %q", fnCmd.Use, "fn")
	}
}

func TestFnCmd_Aliases(t *testing.T) {
	aliases := fnCmd.Aliases
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

func TestFnCmd_PathFlag(t *testing.T) {
	flag := fnCmd.Flags().Lookup("path")
	if flag == nil {
		t.Fatal("Expected 'path' flag to be defined")
	}

	if flag.Shorthand != "p" {
		t.Errorf("path flag shorthand = %q, want %q", flag.Shorthand, "p")
	}
}
