package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCmd(t *testing.T) {
	// Test that rootCmd is properly initialized
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	if rootCmd.Use != "gw" {
		t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "gw")
	}
}

func TestRootCmd_HasSubcommands(t *testing.T) {
	expectedCommands := []string{"add", "ls", "rm", "sw", "exec", "init", "fd"}

	commands := rootCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
		// Also check by name without args
		for _, name := range expectedCommands {
			if cmd.Name() == name {
				commandMap[name] = true
			}
		}
	}

	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range commands {
			if cmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command %q not found in rootCmd", expected)
		}
	}
}

func TestRootCmd_Version(t *testing.T) {
	// Version should be set (default is "dev")
	if rootCmd.Version == "" {
		t.Error("rootCmd.Version should not be empty")
	}
}

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	return buf.String(), err
}

func TestRootCmd_Help(t *testing.T) {
	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output == "" {
		t.Error("Expected help output, got empty string")
	}
}
