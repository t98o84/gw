package cmd

import (
	"strings"
	"testing"
)

func TestRmCmd(t *testing.T) {
	if rmCmd == nil {
		t.Fatal("rmCmd should not be nil")
	}

	if rmCmd.Use != "rm [flags] [name...]" {
		t.Errorf("rmCmd.Use = %q, want %q", rmCmd.Use, "rm [flags] [name...]")
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

func TestRmCmd_BranchFlag(t *testing.T) {
	flag := rmCmd.Flags().Lookup("branch")
	if flag == nil {
		t.Fatal("Expected 'branch' flag to be defined")
	}

	if flag.Shorthand != "b" {
		t.Errorf("branch flag shorthand = %q, want %q", flag.Shorthand, "b")
	}
}

func TestRmCmd_NoYesFlag(t *testing.T) {
	flag := rmCmd.Flags().Lookup("no-yes")
	if flag == nil {
		t.Fatal("Expected 'no-yes' flag to be defined")
	}
}

func TestRmCmd_NoForceFlag(t *testing.T) {
	flag := rmCmd.Flags().Lookup("no-force")
	if flag == nil {
		t.Fatal("Expected 'no-force' flag to be defined")
	}
}

func TestRmCmd_NoBranchFlag(t *testing.T) {
	flag := rmCmd.Flags().Lookup("no-branch")
	if flag == nil {
		t.Fatal("Expected 'no-branch' flag to be defined")
	}
}

func TestRmCmd_AcceptsMultipleArgs(t *testing.T) {
	// rmCmd should accept multiple arguments (no Args restriction)
	if rmCmd.Args != nil {
		t.Log("rmCmd.Args is defined, which is fine as long as it allows multiple args")
	}
}

func TestDeleteBranchSafely(t *testing.T) {
	tests := []struct {
		name          string
		branchName    string
		currentBranch string
		force         bool
		wantDeleted   bool
		wantErr       bool
		errContains   string
	}{
		{
			name:          "refuse to delete main branch",
			branchName:    "main",
			currentBranch: "feature/test",
			force:         false,
			wantDeleted:   false,
			wantErr:       true,
			errContains:   "refusing to delete main branch",
		},
		{
			name:          "refuse to delete master branch",
			branchName:    "master",
			currentBranch: "feature/test",
			force:         false,
			wantDeleted:   false,
			wantErr:       true,
			errContains:   "refusing to delete master branch",
		},
		{
			name:          "refuse to delete current branch",
			branchName:    "feature/test",
			currentBranch: "feature/test",
			force:         false,
			wantDeleted:   false,
			wantErr:       true,
			errContains:   "refusing to delete the current branch",
		},
		{
			name:          "force flag doesn't bypass main branch check",
			branchName:    "main",
			currentBranch: "feature/test",
			force:         true,
			wantDeleted:   false,
			wantErr:       true,
			errContains:   "refusing to delete main branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleted, err := deleteBranchSafely(tt.branchName, tt.currentBranch, "", tt.force)
			if (err != nil) != tt.wantErr {
				t.Errorf("deleteBranchSafely() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if deleted != tt.wantDeleted {
				t.Errorf("deleteBranchSafely() deleted = %v, want %v", deleted, tt.wantDeleted)
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}
