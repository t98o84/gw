package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These tests exercise the *behaviour* of the generated shell wrappers rather
// than asserting on substrings of the generated script. A stub `gw` is placed
// first on PATH so that `command gw ...` inside the wrapper resolves to it: the
// stub emulates `gw close --print-path` output and records the arguments the
// wrapper ultimately passes to `gw rm`. This is what actually guards issue #34
// (the worktree path and the -y flag must reach `gw rm` as separate arguments).

// stubGw is a bash script (run via its shebang regardless of the calling shell)
// that emulates the two gw sub-commands the wrapper invokes. It mirrors the real
// 'gw close --print-path' protocol: stdout = main path, stderr line 1 = worktree
// path, line 2 = force flag (-y), line 3 = branch flag (-b).
const stubGw = `#!/usr/bin/env bash
case "$1" in
  close)
    echo "$GW_MAIN_PATH"
    echo "$GW_WT_PATH" >&2
    force=""
    branch=""
    for a in "$@"; do
      case "$a" in
        -y|-f|--yes|--force) force="-y" ;;
        -b|--branch) branch="-b" ;;
      esac
    done
    if [ "$GW_STUB_FORCE" = "1" ] && [ -z "$force" ]; then force="-y"; fi
    for a in "$@"; do
      case "$a" in --no-yes|--no-force) force="" ;; esac
    done
    echo "$force" >&2
    echo "$branch" >&2
    ;;
  rm)
    shift
    : > "$GW_RM_ARGS_FILE"
    for a in "$@"; do printf '%s\000' "$a" >> "$GW_RM_ARGS_FILE"; done
    ;;
esac
`

// runWrapperClose sources the given wrapper script in the given shell, runs
// `gw <closeArgs>`, and returns the arguments the wrapper passed to `gw rm`.
func runWrapperClose(t *testing.T, shell, wrapper, wrapperExt string, stubForce bool, closeArgs ...string) []string {
	t.Helper()

	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "gw"), []byte(stubGw), 0o755); err != nil {
		t.Fatalf("write stub gw: %v", err)
	}

	mainPath := filepath.Join(dir, "main") // must exist: wrapper does `cd "$main"`
	wtPath := filepath.Join(dir, "wt")     // recorded, not removed (rm is stubbed)
	for _, p := range []string{mainPath, wtPath} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
	}

	wrapperFile := filepath.Join(dir, "wrapper"+wrapperExt)
	if err := os.WriteFile(wrapperFile, []byte(wrapper), 0o644); err != nil {
		t.Fatalf("write wrapper: %v", err)
	}
	rmArgsFile := filepath.Join(dir, "rm_args")

	script := "source '" + wrapperFile + "'; gw " + strings.Join(closeArgs, " ")
	cmd := exec.Command(shell, "-c", script)
	force := "0"
	if stubForce {
		force = "1"
	}
	cmd.Env = append(os.Environ(),
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"GW_MAIN_PATH="+mainPath,
		"GW_WT_PATH="+wtPath,
		"GW_RM_ARGS_FILE="+rmArgsFile,
		"GW_STUB_FORCE="+force,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("running wrapper in %s failed: %v\nstderr: %s", shell, err, stderr.String())
	}

	data, err := os.ReadFile(rmArgsFile)
	if err != nil {
		t.Fatalf("gw rm was never called (no args file): %v\nstderr: %s", err, stderr.String())
	}
	// Split on the NUL separator; drop the trailing empty element.
	parts := strings.Split(string(data), "\x00")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	// Normalise the recorded worktree path to the placeholder "WT" so callers
	// can write expectations without knowing the temp dir.
	for i, p := range parts {
		if p == wtPath {
			parts[i] = "WT"
		}
	}
	return parts
}

func requireShell(t *testing.T, shell string) {
	t.Helper()
	if _, err := exec.LookPath(shell); err != nil {
		t.Skipf("%s not available: %v", shell, err)
	}
}

func equalArgs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestCloseWrapperBehavior_PosixShells(t *testing.T) {
	cases := []struct {
		name      string
		stubForce bool
		closeArgs []string
		wantRm    []string
	}{
		// #34 core: with force enabled the path and -y must reach gw rm as two
		// separate args (the bug passed "<path>\n-y" as one arg).
		{"config force separates path and -y", true, []string{"close"}, []string{"-y", "WT"}},
		// Ancillary: `gw close -y` must forward the CLI flag, not only config.
		{"cli -y is forwarded", false, []string{"close", "-y"}, []string{"-y", "WT"}},
		// No force: gw rm is called without -y.
		{"no force omits -y", false, []string{"close"}, []string{"WT"}},
		// #36: -b is forwarded to gw rm so the branch is also deleted.
		{"cli -b forwards branch deletion", false, []string{"close", "-b"}, []string{"-b", "WT"}},
		// #36: -f -b forwards both force and branch (force-delete unmerged branch).
		{"force and branch forwarded together", false, []string{"close", "-f", "-b"}, []string{"-y", "-b", "WT"}},
	}

	for _, shell := range []string{"bash", "zsh"} {
		shell := shell
		t.Run(shell, func(t *testing.T) {
			requireShell(t, shell)
			for _, tc := range cases {
				tc := tc
				t.Run(tc.name, func(t *testing.T) {
					got := runWrapperClose(t, shell, bashZshInit, ".sh", tc.stubForce, tc.closeArgs...)
					if !equalArgs(got, tc.wantRm) {
						t.Errorf("gw rm args = %q, want %q", got, tc.wantRm)
					}
				})
			}
		})
	}
}

func TestCloseWrapperBehavior_Fish(t *testing.T) {
	requireShell(t, "fish")
	cases := []struct {
		name      string
		stubForce bool
		closeArgs []string
		wantRm    []string
	}{
		{"config force separates path and -y", true, []string{"close"}, []string{"-y", "WT"}},
		{"cli -y is forwarded", false, []string{"close", "-y"}, []string{"-y", "WT"}},
		{"no force omits -y", false, []string{"close"}, []string{"WT"}},
		{"cli -b forwards branch deletion", false, []string{"close", "-b"}, []string{"-b", "WT"}},
		{"force and branch forwarded together", false, []string{"close", "-f", "-b"}, []string{"-y", "-b", "WT"}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := runWrapperClose(t, "fish", fishInit, ".fish", tc.stubForce, tc.closeArgs...)
			if !equalArgs(got, tc.wantRm) {
				t.Errorf("gw rm args = %q, want %q", got, tc.wantRm)
			}
		})
	}
}
