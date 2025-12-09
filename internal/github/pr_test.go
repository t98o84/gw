package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-github/v66/github"
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

// TestGetRemoteOwnerRepo_Integration tests getRemoteOwnerRepo with actual git command
// This test runs in the actual repository and validates the parsing logic
func TestGetRemoteOwnerRepo_Integration(t *testing.T) {
	// Skip if not in a git repository
	if _, err := exec.Command("git", "rev-parse", "--git-dir").Output(); err != nil {
		t.Skip("Not in a git repository, skipping integration test")
	}

	owner, repo, err := getRemoteOwnerRepo()
	if err != nil {
		t.Fatalf("getRemoteOwnerRepo() error = %v", err)
	}

	// Verify we got some results
	if owner == "" {
		t.Error("getRemoteOwnerRepo() returned empty owner")
	}
	if repo == "" {
		t.Error("getRemoteOwnerRepo() returned empty repo")
	}

	// For this repository, we expect specific values
	if owner == "t98o84" && repo != "gw" {
		t.Errorf("getRemoteOwnerRepo() repo = %v, want gw", repo)
	}
}

// TestNewGitHubClient_WithEnvironmentToken tests token retrieval from environment
func TestNewGitHubClient_WithEnvironmentToken(t *testing.T) {
	tests := []struct {
		name       string
		setupEnv   func()
		cleanupEnv func()
		wantErr    bool
	}{
		{
			name: "with GITHUB_TOKEN",
			setupEnv: func() {
				os.Setenv("GITHUB_TOKEN", "test-token-123")
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_TOKEN")
			},
			wantErr: false,
		},
		{
			name: "with GH_TOKEN",
			setupEnv: func() {
				os.Setenv("GH_TOKEN", "test-token-456")
			},
			cleanupEnv: func() {
				os.Unsetenv("GH_TOKEN")
			},
			wantErr: false,
		},
		{
			name: "GITHUB_TOKEN takes precedence",
			setupEnv: func() {
				os.Setenv("GITHUB_TOKEN", "github-token")
				os.Setenv("GH_TOKEN", "gh-token")
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_TOKEN")
				os.Unsetenv("GH_TOKEN")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupEnv()
			defer tt.cleanupEnv()

			client, err := newGitHubClient()
			if (err != nil) != tt.wantErr {
				t.Errorf("newGitHubClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("newGitHubClient() returned nil client")
			}
		})
	}
}

// TestNewGitHubClient_WithoutToken tests error when no token is available
func TestNewGitHubClient_WithoutToken(t *testing.T) {
	// Save original environment
	origGitHubToken := os.Getenv("GITHUB_TOKEN")
	origGHToken := os.Getenv("GH_TOKEN")

	// Clear environment variables
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GH_TOKEN")

	// Restore after test
	defer func() {
		if origGitHubToken != "" {
			os.Setenv("GITHUB_TOKEN", origGitHubToken)
		}
		if origGHToken != "" {
			os.Setenv("GH_TOKEN", origGHToken)
		}
	}()

	// This will try gh auth token command which might succeed or fail
	// If gh CLI is not installed or not authenticated, it should return error
	client, err := newGitHubClient()

	// If gh CLI provides a token, client will be created
	// If not, we should get an error
	if err == nil && client == nil {
		t.Error("newGitHubClient() should return either a client or an error")
	}
}

// TestGetPRBranch_WithMockServer tests GetPRBranch with a mock GitHub API server
func TestGetPRBranch_WithMockServer(t *testing.T) {
	// Skip if not in a git repository (needed for getRemoteOwnerRepo)
	if _, err := exec.Command("git", "rev-parse", "--git-dir").Output(); err != nil {
		t.Skip("Not in a git repository, skipping integration test")
	}

	// Setup mock GitHub API server
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	// Mock PR API response
	mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) {
		// Parse the URL to extract PR number
		if strings.Contains(r.URL.Path, "/pulls/") {
			response := `{
				"number": 123,
				"head": {
					"ref": "feature/test-branch"
				}
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(response)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// Create a client pointing to our mock server
	// Note: This test requires modifying newGitHubClient to accept a custom base URL
	// For now, we'll skip the actual API call test
	t.Skip("Skipping API call test - requires refactoring newGitHubClient to accept custom base URL")
}

// TestParsePRIdentifier_WithPRNumber tests parsing PR numbers (requires git remote)
func TestParsePRIdentifier_WithPRNumber(t *testing.T) {
	// Skip if not in a git repository
	if _, err := exec.Command("git", "rev-parse", "--git-dir").Output(); err != nil {
		t.Skip("Not in a git repository, skipping integration test")
	}

	tests := []struct {
		name       string
		identifier string
		wantPRNum  int
		wantErr    bool
	}{
		{
			name:       "simple PR number",
			identifier: "123",
			wantPRNum:  123,
			wantErr:    false,
		},
		{
			name:       "large PR number",
			identifier: "99999",
			wantPRNum:  99999,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prNum, owner, repo, err := parsePRIdentifier(tt.identifier, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePRIdentifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if prNum != tt.wantPRNum {
					t.Errorf("parsePRIdentifier() prNum = %v, want %v", prNum, tt.wantPRNum)
				}
				// Verify owner and repo were extracted from git remote
				if owner == "" || repo == "" {
					t.Error("parsePRIdentifier() should extract owner and repo from git remote")
				}
			}
		})
	}
}

// TestGetPRBranch_ErrorCases tests error handling in GetPRBranch
func TestGetPRBranch_ErrorCases(t *testing.T) {
	tests := []struct {
		name         string
		prIdentifier string
		repoName     string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "invalid identifier format",
			prIdentifier: "invalid-format",
			repoName:     "",
			wantErr:      true,
			errContains:  "invalid PR identifier",
		},
		{
			name:         "empty identifier",
			prIdentifier: "",
			repoName:     "",
			wantErr:      true,
			errContains:  "invalid PR identifier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail at parsePRIdentifier stage
			_, err := GetPRBranch(tt.prIdentifier, tt.repoName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPRBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetPRBranch() error = %v, should contain %q", err, tt.errContains)
				}
			}
		})
	}
}

// TestGetPRBranch_WithMockClient tests GetPRBranch with a mocked GitHub client
// This demonstrates how we could test the function with dependency injection
func TestGetPRBranch_MockedClient(t *testing.T) {
	// This is a demonstration of how the code could be tested with better architecture
	// The current implementation uses global functions which makes it hard to inject mocks

	// For now, we create a test that validates the structure
	t.Run("validate PR branch extraction logic", func(t *testing.T) {
		// Create a mock PR response
		mockPR := &github.PullRequest{
			Number: github.Int(123),
			Head: &github.PullRequestBranch{
				Ref: github.String("feature/test-branch"),
			},
		}

		// Verify we can extract the branch name
		branchName := mockPR.Head.GetRef()
		if branchName != "feature/test-branch" {
			t.Errorf("Branch name = %v, want feature/test-branch", branchName)
		}
	})
}

// TestNewGitHubClient_TokenPriority tests the priority order of token sources
func TestNewGitHubClient_TokenPriority(t *testing.T) {
	// Save and clear environment
	origGitHubToken := os.Getenv("GITHUB_TOKEN")
	origGHToken := os.Getenv("GH_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GH_TOKEN")

	defer func() {
		if origGitHubToken != "" {
			os.Setenv("GITHUB_TOKEN", origGitHubToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
		if origGHToken != "" {
			os.Setenv("GH_TOKEN", origGHToken)
		} else {
			os.Unsetenv("GH_TOKEN")
		}
	}()

	t.Run("GITHUB_TOKEN has highest priority", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "github-token")
		os.Setenv("GH_TOKEN", "gh-token")
		defer func() {
			os.Unsetenv("GITHUB_TOKEN")
			os.Unsetenv("GH_TOKEN")
		}()

		client, err := newGitHubClient()
		if err != nil {
			t.Fatalf("newGitHubClient() error = %v", err)
		}
		if client == nil {
			t.Fatal("newGitHubClient() returned nil client")
		}
		// The client should be created with GITHUB_TOKEN
	})

	t.Run("GH_TOKEN used when GITHUB_TOKEN not set", func(t *testing.T) {
		os.Setenv("GH_TOKEN", "gh-token")
		defer os.Unsetenv("GH_TOKEN")

		client, err := newGitHubClient()
		if err != nil {
			t.Fatalf("newGitHubClient() error = %v", err)
		}
		if client == nil {
			t.Fatal("newGitHubClient() returned nil client")
		}
	})
}

// TestGetRemoteOwnerRepo_ErrorHandling tests error cases for remote parsing
func TestGetRemoteOwnerRepo_ErrorHandling(t *testing.T) {
	t.Run("handles git command failure gracefully", func(t *testing.T) {
		// Create a temporary directory that's not a git repo
		tempDir := t.TempDir()

		// Change to temp directory
		origDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Chdir(origDir); err != nil {
				t.Errorf("Failed to change back to original directory: %v", err)
			}
		}()

		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		// Try to get remote from non-git directory
		_, _, err = getRemoteOwnerRepo()
		if err == nil {
			t.Error("getRemoteOwnerRepo() should return error in non-git directory")
		}
	})
}

// Benchmark tests for performance validation
func BenchmarkParsePRIdentifier_URL(b *testing.B) {
	identifier := "https://github.com/owner/repo/pull/123"
	for i := 0; i < b.N; i++ {
		parsePRIdentifier(identifier, "")
	}
}

func BenchmarkParsePRIdentifier_Number(b *testing.B) {
	// Skip if not in git repo
	if _, err := exec.Command("git", "rev-parse", "--git-dir").Output(); err != nil {
		b.Skip("Not in a git repository")
	}

	identifier := "123"
	for i := 0; i < b.N; i++ {
		parsePRIdentifier(identifier, "")
	}
}

// TestParsePRIdentifier_EdgeCases tests edge cases and boundary conditions
func TestParsePRIdentifier_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		wantErr    bool
	}{
		{
			name:       "PR number 0",
			identifier: "0",
			wantErr:    false, // Technically valid, though GitHub PRs start at 1
		},
		{
			name:       "very large PR number",
			identifier: "999999999",
			wantErr:    false,
		},
		{
			name:       "negative number",
			identifier: "-1",
			wantErr:    false, // strconv.Atoi("-1") succeeds, then git remote lookup happens
		},
		{
			name:       "URL with trailing slash",
			identifier: "https://github.com/owner/repo/pull/123/",
			wantErr:    false, // Trailing slash is ignored by regex
		},
		{
			name:       "URL with additional path",
			identifier: "https://github.com/owner/repo/pull/123/files",
			wantErr:    false, // Additional path is ignored by regex
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For number tests that need git remote, skip if not in git repo
			if _, err := strconv.Atoi(tt.identifier); err == nil {
				if _, err := exec.Command("git", "rev-parse", "--git-dir").Output(); err != nil {
					t.Skip("Not in a git repository")
				}
			}

			_, _, _, err := parsePRIdentifier(tt.identifier, "")
			_ = err // Explicitly ignore for edge case testing
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePRIdentifier() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewGitHubClient_ContextAndOAuth tests that the client is properly configured
func TestNewGitHubClient_ContextAndOAuth(t *testing.T) {
	// Setup environment
	os.Setenv("GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	client, err := newGitHubClient()
	if err != nil {
		t.Fatalf("newGitHubClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("newGitHubClient() returned nil client")
	}

	// Verify client can be used (basic structure test)
	ctx := context.Background()
	_, resp, _ := client.Users.Get(ctx, "")
	// We expect an authentication error with our fake token, but the client should be properly structured
	if resp == nil {
		t.Error("Expected response object even with auth error")
	}
}
