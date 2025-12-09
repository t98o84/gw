package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

// GetPRBranch extracts the branch name from a PR number or URL
func GetPRBranch(prIdentifier string, repoName string) (string, error) {
	prNumber, owner, repo, err := parsePRIdentifier(prIdentifier, repoName)
	if err != nil {
		return "", err
	}

	client, err := newGitHubClient()
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get PR #%d: %w", prNumber, err)
	}

	return pr.Head.GetRef(), nil
}

// parsePRIdentifier parses a PR number or URL and returns PR number, owner, and repo
func parsePRIdentifier(identifier string, defaultRepo string) (int, string, string, error) {
	// Try to parse as URL
	// Formats:
	//   https://github.com/owner/repo/pull/123
	//   github.com/owner/repo/pull/123
	urlRegex := regexp.MustCompile(`(?:https?://)?github\.com/([^/]+)/([^/]+)/pull/(\d+)`)
	if matches := urlRegex.FindStringSubmatch(identifier); matches != nil {
		prNum, _ := strconv.Atoi(matches[3])
		return prNum, matches[1], matches[2], nil
	}

	// Try to parse as number
	prNum, err := strconv.Atoi(identifier)
	if err == nil {
		// Need to get owner/repo from git remote
		owner, repo, err := getRemoteOwnerRepo()
		if err != nil {
			return 0, "", "", err
		}
		return prNum, owner, repo, nil
	}

	return 0, "", "", fmt.Errorf("invalid PR identifier: %s (use PR number or URL)", identifier)
}

// getRemoteOwnerRepo extracts owner and repo from git remote origin
func getRemoteOwnerRepo() (string, string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	url := strings.TrimSpace(string(out))

	// Parse SSH format: git@github.com:owner/repo.git or custom.github.com:owner/repo.git
	sshRegex := regexp.MustCompile(`(?:git@)?(?:[^:]+\.)?github\.com:([^/]+)/([^/]+?)(?:\.git)?$`)
	if matches := sshRegex.FindStringSubmatch(url); matches != nil {
		return matches[1], matches[2], nil
	}

	// Parse HTTPS format: https://github.com/owner/repo.git
	httpsRegex := regexp.MustCompile(`https://(?:[^/]+\.)?github\.com/([^/]+)/([^/]+?)(?:\.git)?$`)
	if matches := httpsRegex.FindStringSubmatch(url); matches != nil {
		return matches[1], matches[2], nil
	}

	return "", "", fmt.Errorf("failed to parse remote URL: %s", url)
}

// newGitHubClient creates a new GitHub client with authentication
func newGitHubClient() (*github.Client, error) {
	// Try to get token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}

	// Try to get token from gh CLI
	if token == "" {
		cmd := exec.Command("gh", "auth", "token")
		out, err := cmd.Output()
		if err == nil {
			token = strings.TrimSpace(string(out))
		}
	}

	if token == "" {
		return nil, fmt.Errorf("GitHub token not found. Set GITHUB_TOKEN or GH_TOKEN environment variable, or login with 'gh auth login'")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc), nil
}
