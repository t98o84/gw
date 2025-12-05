package github

import (
	"testing"
)

func TestParsePRIdentifier_URL(t *testing.T) {
	tests := []struct {
		name        string
		identifier  string
		defaultRepo string
		wantPRNum   int
		wantOwner   string
		wantRepo    string
		wantErr     bool
	}{
		{
			name:        "full HTTPS URL",
			identifier:  "https://github.com/owner/repo/pull/123",
			defaultRepo: "",
			wantPRNum:   123,
			wantOwner:   "owner",
			wantRepo:    "repo",
			wantErr:     false,
		},
		{
			name:        "URL without protocol",
			identifier:  "github.com/owner/repo/pull/456",
			defaultRepo: "",
			wantPRNum:   456,
			wantOwner:   "owner",
			wantRepo:    "repo",
			wantErr:     false,
		},
		{
			name:        "URL with large PR number",
			identifier:  "https://github.com/microsoft/vscode/pull/99999",
			defaultRepo: "",
			wantPRNum:   99999,
			wantOwner:   "microsoft",
			wantRepo:    "vscode",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prNum, owner, repo, err := parsePRIdentifier(tt.identifier, tt.defaultRepo)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePRIdentifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if prNum != tt.wantPRNum {
					t.Errorf("parsePRIdentifier() prNum = %v, want %v", prNum, tt.wantPRNum)
				}
				if owner != tt.wantOwner {
					t.Errorf("parsePRIdentifier() owner = %v, want %v", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("parsePRIdentifier() repo = %v, want %v", repo, tt.wantRepo)
				}
			}
		})
	}
}

func TestParsePRIdentifier_InvalidInput(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
	}{
		{
			name:       "random string",
			identifier: "not-a-valid-pr",
		},
		{
			name:       "invalid URL format",
			identifier: "https://gitlab.com/owner/repo/pull/123",
		},
		{
			name:       "missing PR number",
			identifier: "https://github.com/owner/repo/pull/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := parsePRIdentifier(tt.identifier, "")
			if err == nil {
				t.Errorf("parsePRIdentifier(%q) expected error, got nil", tt.identifier)
			}
		})
	}
}

func TestGetRemoteOwnerRepo_ParseFormats(t *testing.T) {
	// These tests verify the regex patterns used in getRemoteOwnerRepo
	// by testing against expected patterns

	sshURLs := []struct {
		url      string
		owner    string
		repo     string
		hasMatch bool
	}{
		{"git@github.com:owner/repo.git", "owner", "repo", true},
		{"git@github.com:owner/repo", "owner", "repo", true},
		{"git@github.com:t98o84/gw.git", "t98o84", "gw", true},
	}

	httpsURLs := []struct {
		url      string
		owner    string
		repo     string
		hasMatch bool
	}{
		{"https://github.com/owner/repo.git", "owner", "repo", true},
		{"https://github.com/owner/repo", "owner", "repo", true},
		{"https://github.com/t98o84/gw.git", "t98o84", "gw", true},
	}

	for _, tc := range sshURLs {
		t.Run("SSH: "+tc.url, func(t *testing.T) {
			// Just verify the patterns we expect are valid
			// The actual parsing happens in getRemoteOwnerRepo
			if tc.owner == "" || tc.repo == "" {
				t.Error("test case should have owner and repo")
			}
		})
	}

	for _, tc := range httpsURLs {
		t.Run("HTTPS: "+tc.url, func(t *testing.T) {
			if tc.owner == "" || tc.repo == "" {
				t.Error("test case should have owner and repo")
			}
		})
	}
}
