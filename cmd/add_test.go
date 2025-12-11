package cmd

import (
	"testing"
)

func TestAddCmd(t *testing.T) {
	if addCmd == nil {
		t.Fatal("addCmd should not be nil")
	}

	if addCmd.Use != "add <branch>" {
		t.Errorf("addCmd.Use = %q, want %q", addCmd.Use, "add <branch>")
	}
}

func TestAddCmd_Aliases(t *testing.T) {
	aliases := addCmd.Aliases
	expected := []string{"a"}

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

func TestAddCmd_BranchFlag(t *testing.T) {
	flag := addCmd.Flags().Lookup("branch")
	if flag == nil {
		t.Fatal("Expected 'branch' flag to be defined")
	}

	if flag.Shorthand != "b" {
		t.Errorf("branch flag shorthand = %q, want %q", flag.Shorthand, "b")
	}
}

func TestAddCmd_PRFlag(t *testing.T) {
	flag := addCmd.Flags().Lookup("pr")
	if flag == nil {
		t.Fatal("Expected 'pr' flag to be defined")
	}

	if flag.Shorthand != "p" {
		t.Errorf("pr flag shorthand = %q, want %q", flag.Shorthand, "p")
	}
}

func TestAddCmd_OpenFlag(t *testing.T) {
	flag := addCmd.Flags().Lookup("open")
	if flag == nil {
		t.Fatal("Expected 'open' flag to be defined")
	}
}

func TestAddCmd_EditorFlag(t *testing.T) {
	flag := addCmd.Flags().Lookup("editor")
	if flag == nil {
		t.Fatal("Expected 'editor' flag to be defined")
	}

	if flag.Shorthand != "e" {
		t.Errorf("editor flag shorthand = %q, want %q", flag.Shorthand, "e")
	}
}
