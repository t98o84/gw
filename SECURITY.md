# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Currently supported versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.3.x   | :white_check_mark: |
| < 0.3   | :x:                |

## Reporting a Vulnerability

The gw team takes security issues seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report a Security Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

1. **GitHub Security Advisory** (Preferred)
   - Go to the [Security tab](https://github.com/t98o84/gw/security/advisories) of this repository
   - Click "Report a vulnerability"
   - Fill out the form with details about the vulnerability

2. **Email**
   - Send an email to the repository owner through GitHub
   - Include "SECURITY" in the subject line
   - Provide a detailed description of the vulnerability

### What to Include in Your Report

Please include the following information to help us better understand the nature and scope of the issue:

- Type of issue (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
- Full paths of source file(s) related to the manifestation of the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### Response Timeline

- **Initial Response**: We will acknowledge your report within 48 hours
- **Status Update**: We will provide a more detailed response within 7 days, indicating the next steps
- **Fix Timeline**: We will work to release a fix as quickly as possible, depending on the complexity

### Disclosure Policy

- We will coordinate the disclosure timeline with you
- Once a fix is available, we will:
  1. Release a new version with the fix
  2. Publish a security advisory on GitHub
  3. Credit you (if you wish) in the security advisory and release notes

### Bug Bounty

This project does not currently have a bug bounty program. However, we deeply appreciate responsible disclosure and will publicly acknowledge your contribution (with your permission).

## Security Best Practices for Users

### Keep gw Updated

Always use the latest version of gw to ensure you have the latest security patches:

```bash
# Homebrew
brew upgrade gw

# Go
go install github.com/t98o84/gw@latest
```

### Configuration File Security

- Keep your configuration files (`~/.config/gw/config.yaml` and `gw.yaml`) secure
- Be careful with hook commands that might execute untrusted input
- Review hook commands in `gw.yaml` before using worktrees from untrusted sources

### GitHub Token Security

- Store your GitHub token securely (use `gh auth login` when possible)
- Never commit tokens to version control
- Use tokens with minimal required permissions
- Regularly rotate your GitHub tokens

### Hook Command Safety

When using hooks in `gw.yaml`:
- Avoid executing untrusted scripts or commands
- Validate input before using in hooks
- Be cautious with commands that modify files or execute with elevated privileges
- Review hooks in repositories you don't maintain

## Known Security Considerations

### Shell Command Execution

gw executes shell commands for:
- Git operations
- Hook commands
- Editor launching

Users should be aware that:
- Hook commands in `gw.yaml` are executed with the user's permissions
- Malicious `gw.yaml` files could execute harmful commands
- Always review `gw.yaml` in repositories you don't control

### GitHub API Access

- gw uses GitHub tokens for API access
- Tokens are read from environment variables or `gh` CLI
- Never share your GitHub token with others

## Comments on This Policy

If you have suggestions on how this process could be improved, please submit a pull request.
