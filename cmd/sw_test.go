package cmd

import (
	"testing"
)

func TestSwCmd(t *testing.T) {
	if swCmd == nil {
		t.Fatal("swCmd should not be nil")
	}

	if swCmd.Use != "sw [flags] [name]" {
		t.Errorf("swCmd.Use = %q, want %q", swCmd.Use, "sw [flags] [name]")
	}
}

func TestSwCmd_Aliases(t *testing.T) {
	aliases := swCmd.Aliases
	expected := []string{"s"}

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

func TestSwCmd_PrintPathFlag(t *testing.T) {
	flag := swCmd.Flags().Lookup("print-path")
	if flag == nil {
		t.Fatal("Expected 'print-path' flag to be defined")
	}
}
