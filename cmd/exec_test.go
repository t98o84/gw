package cmd

import (
	"testing"
)

func TestExecCmd(t *testing.T) {
	if execCmd == nil {
		t.Fatal("execCmd should not be nil")
	}

	if execCmd.Use != "exec <name> <command...>" {
		t.Errorf("execCmd.Use = %q, want %q", execCmd.Use, "exec <name> <command...>")
	}
}

func TestExecCmd_Aliases(t *testing.T) {
	aliases := execCmd.Aliases
	expected := []string{"e"}

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

func TestExecCmd_RequiresArgs(t *testing.T) {
	// execCmd requires at least 1 argument (the command to run)
	// The Args field should be set to MinimumNArgs(1)
	if execCmd.Args == nil {
		t.Error("execCmd.Args should be defined")
	}
}
