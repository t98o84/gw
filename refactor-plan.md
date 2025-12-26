# Phase 1 Implementation Plan

## 1. Implementation Scope

### Covered Items
- **Introduction of shell abstraction layer**: Abstraction of direct `exec.Command` calls for testability
- **Unification of fzf integration**: Consolidate fzf calls scattered across 3 locations into a unified interface
- **Split runAdd function**: Split the complex 96-line function by responsibility to improve maintainability

### Items Not Covered (Phase 2 and Beyond)
- Improvement of global variable flag management (Phase 2)
- Implementation of unified error handling wrapper (Phase 2)
- Integration tests with GitHub API (Phase 3)
- Performance optimization (Phase 3)

### Completion Criteria
- [ ] `internal/shell/executor.go` is implemented, and all 15 `exec.Command` calls are migrated
- [ ] `internal/fzf/selector.go` is implemented, and 3 fzf calls are migrated to unified interface
- [ ] Complexity of `cmd/add.go:runAdd` reduced to 10 or below
- [ ] Test coverage of new packages (`internal/shell`, `internal/fzf`) is 80% or above
- [ ] All existing tests continue to pass
- [ ] Build and test execution successful in Docker environment

## 2. Architecture Design

### Package Structure

```
gw/
├── cmd/                        # CLI command layer
│   ├── add.go                 # Refactored: Uses Executor/FzfSelector
│   ├── sw.go                  # Refactored: Uses FzfSelector
│   ├── rm.go                  # Refactored: Uses FzfSelector
│   ├── fzf.go                 # Deletion candidate: Integrate into fzf package
│   └── ...
├── internal/
│   ├── shell/                 # New: Command execution abstraction
│   │   ├── executor.go        # interface + implementation
│   │   └── executor_test.go   # Tests including mock
│   ├── fzf/                   # New: fzf integration
│   │   ├── selector.go        # interface + implementation
│   │   └── selector_test.go   # Tests including mock
│   ├── git/
│   │   ├── worktree.go        # Refactored: Uses Executor
│   │   ├── naming.go          # No change
│   │   └── ...
│   └── github/
│       └── pr.go              # Refactored: Uses Executor
```

### Dependencies

```
cmd/* 
  ↓ (uses)
internal/shell/Executor ←┐
internal/fzf/Selector    │
  ↓ (uses)               │
internal/git/*           │
internal/github/*        │
  ↓ (uses)               │
internal/shell/Executor ─┘
```

**Design Principles**:
- `cmd` package depends only on `internal/shell` and `internal/fzf` interfaces
- `internal/git` and `internal/github` receive `internal/shell.Executor` via DI
- Inject `MockExecutor` during tests to simulate command execution

## 3. Detailed Design

### 3.1. Shell Abstraction Layer

**File**: `internal/shell/executor.go`

```go
package shell

import (
    "io"
    "os/exec"
)

// Executor abstracts the execution of shell commands
type Executor interface {
    // Execute runs a command and returns its standard output
    Execute(name string, args ...string) ([]byte, error)
    
    // ExecuteWithStdio runs a command with standard I/O connected
    ExecuteWithStdio(name string, args ...string) error
    
    // LookPath checks if a command exists
    LookPath(name string) (string, error)
}

// RealExecutor is an implementation that performs actual command execution
type RealExecutor struct{}

func NewRealExecutor() *RealExecutor {
    return &RealExecutor{}
}

func (e *RealExecutor) Execute(name string, args ...string) ([]byte, error) {
    cmd := exec.Command(name, args...)
    return cmd.Output()
}

func (e *RealExecutor) ExecuteWithStdio(name string, args ...string) error {
    cmd := exec.Command(name, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin
    return cmd.Run()
}

func (e *RealExecutor) LookPath(name string) (string, error) {
    return exec.LookPath(name)
}

// MockExecutor is a mock implementation for testing
type MockExecutor struct {
    ExecuteFunc         func(string, ...string) ([]byte, error)
    ExecuteWithStdioFunc func(string, ...string) error
    LookPathFunc        func(string) (string, error)
}

func (m *MockExecutor) Execute(name string, args ...string) ([]byte, error) {
    if m.ExecuteFunc != nil {
        return m.ExecuteFunc(name, args...)
    }
    return []byte{}, nil
}

func (m *MockExecutor) ExecuteWithStdio(name string, args ...string) error {
    if m.ExecuteWithStdioFunc != nil {
        return m.ExecuteWithStdioFunc(name, args...)
    }
    return nil
}

func (m *MockExecutor) LookPath(name string) (string, error) {
    if m.LookPathFunc != nil {
        return m.LookPathFunc(name)
    }
    return "/usr/bin/" + name, nil
}
```

**Migration Strategy**:
1. DI `Executor` into all functions in `internal/git/worktree.go`
2. DI `Executor` into `GetPRBranch` in `internal/github/pr.go`
3. Initialize `RealExecutor` in `cmd/*` and pass it

### 3.2. Unification of fzf Integration

**File**: `internal/fzf/selector.go`

```go
package fzf

import (
    "fmt"
    "os"
    "strings"
    
    "github.com/t98o84/gw/internal/shell"
)

// SelectOptions are options for fzf selection
type SelectOptions struct {
    Items       []string       // List of choices
    Prompt      string         // Prompt message
    Multi       bool           // Allow multiple selection
    Height      string         // Display height (default: "40%")
    Reverse     bool           // Reverse order display
}

// Selector abstracts interactive selection with fzf
type Selector interface {
    // Select performs single selection
    Select(opts SelectOptions) (string, error)
    
    // SelectMulti performs multiple selection
    SelectMulti(opts SelectOptions) ([]string, error)
    
    // IsAvailable checks if fzf is available
    IsAvailable() bool
}

// FzfSelector is an implementation using the fzf command
type FzfSelector struct {
    executor shell.Executor
}

func NewFzfSelector(executor shell.Executor) *FzfSelector {
    return &FzfSelector{executor: executor}
}

func (s *FzfSelector) IsAvailable() bool {
    _, err := s.executor.LookPath("fzf")
    return err == nil
}

func (s *FzfSelector) Select(opts SelectOptions) (string, error) {
    if !s.IsAvailable() {
        return "", fmt.Errorf("fzf is not installed")
    }
    
    results, err := s.selectInternal(opts, false)
    if err != nil {
        return "", err
    }
    if len(results) == 0 {
        return "", nil
    }
    return results[0], nil
}

func (s *FzfSelector) SelectMulti(opts SelectOptions) ([]string, error) {
    if !s.IsAvailable() {
        return nil, fmt.Errorf("fzf is not installed")
    }
    
    return s.selectInternal(opts, true)
}

func (s *FzfSelector) selectInternal(opts SelectOptions, multi bool) ([]string, error) {
    // Build fzf arguments
    height := opts.Height
    if height == "" {
        height = "40%"
    }
    
    args := []string{"--height=" + height}
    if opts.Reverse {
        args = append(args, "--reverse")
    }
    
    if multi || opts.Multi {
        args = append(args, "--multi")
        if opts.Prompt == "" {
            opts.Prompt = "Select items (Tab to multi-select): "
        }
    } else {
        if opts.Prompt == "" {
            opts.Prompt = "Select item: "
        }
    }
    args = append(args, "--prompt="+opts.Prompt)
    
    // Execute fzf (special handling needed to pass items to stdin)
    cmd := exec.Command("fzf", args...)
    cmd.Stdin = strings.NewReader(strings.Join(opts.Items, "\n"))
    cmd.Stderr = os.Stderr
    
    out, err := cmd.Output()
    if err != nil {
        // Ctrl+C (exit code 130) is treated as cancellation, not error
        if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
            return nil, nil
        }
        return nil, nil // Other errors also treated as cancellation
    }
    
    output := strings.TrimSpace(string(out))
    if output == "" {
        return nil, nil
    }
    
    // Split multi-line results
    results := strings.Split(output, "\n")
    var trimmed []string
    for _, r := range results {
        if t := strings.TrimSpace(r); t != "" {
            trimmed = append(trimmed, t)
        }
    }
    
    return trimmed, nil
}

// MockSelector is a mock implementation for testing
type MockSelector struct {
    SelectFunc      func(SelectOptions) (string, error)
    SelectMultiFunc func(SelectOptions) ([]string, error)
    IsAvailableFunc func() bool
}

func (m *MockSelector) Select(opts SelectOptions) (string, error) {
    if m.SelectFunc != nil {
        return m.SelectFunc(opts)
    }
    return "", nil
}

func (m *MockSelector) SelectMulti(opts SelectOptions) ([]string, error) {
    if m.SelectMultiFunc != nil {
        return m.SelectMultiFunc(opts)
    }
    return nil, nil
}

func (m *MockSelector) IsAvailable() bool {
    if m.IsAvailableFunc != nil {
        return m.IsAvailableFunc()
    }
    return true
}
```

**Migration Strategy**:
1. Rewrite `selectWorktreeWithFzf` in `cmd/fzf.go` to use `fzf.Selector`
2. Rewrite `selectBranchWithFzf` in `cmd/add.go` to use `fzf.Selector`
3. Eventually delete `cmd/fzf.go` and fully migrate functionality to `internal/fzf`

### 3.3. Split runAdd Function

**Structure After Split**:

```go
// cmd/add.go

type addOptions struct {
    branch         string
    createBranch   bool
    prIdentifier   string
    executor       shell.Executor
    fzfSelector    fzf.Selector
}

func runAdd(cmd *cobra.Command, args []string) error {
    executor := shell.NewRealExecutor()
    selector := fzf.NewFzfSelector(executor)
    
    opts := &addOptions{
        createBranch: addCreateBranch,
        prIdentifier: addPRIdentifier,
        executor:     executor,
        fzfSelector:  selector,
    }
    
    // Step 1: Determine branch name
    branch, err := opts.determineBranch(args)
    if err != nil {
        return err
    }
    if branch == "" {
        return nil // User cancelled
    }
    opts.branch = branch
    
    // Step 2: Get repository information
    repoName, err := git.GetRepoName()
    if err != nil {
        return fmt.Errorf("failed to get repository name: %w", err)
    }
    
    // Step 3: Generate worktree path
    wtPath, err := git.WorktreePath(repoName, branch)
    if err != nil {
        return fmt.Errorf("failed to generate worktree path: %w", err)
    }
    
    // Step 4: Check existing worktree
    if err := opts.checkExistingWorktree(branch); err != nil {
        return err
    }
    
    // Step 5: Ensure branch exists and fetch
    if err := opts.ensureBranchExists(); err != nil {
        return err
    }
    
    // Step 6: Create worktree
    fmt.Printf("Creating worktree at %s for branch %s...\n", wtPath, branch)
    if err := git.AddWithExecutor(executor, wtPath, branch, opts.createBranch); err != nil {
        return err
    }
    
    fmt.Printf("✓ Worktree created: %s\n", wtPath)
    return nil
}

// determineBranch determines the branch name (from PR/fzf/arguments)
func (opts *addOptions) determineBranch(args []string) (string, error) {
    // If PR specified
    if opts.prIdentifier != "" {
        return opts.getBranchFromPR()
    }
    
    // If no arguments, select with fzf
    if len(args) == 0 {
        return opts.selectBranchInteractive()
    }
    
    // If specified by argument
    return args[0], nil
}

// getBranchFromPR gets branch name from PR
func (opts *addOptions) getBranchFromPR() (string, error) {
    repoName, err := git.GetRepoName()
    if err != nil {
        return "", fmt.Errorf("failed to get repository name: %w", err)
    }
    
    branch, err := github.GetPRBranchWithExecutor(opts.executor, opts.prIdentifier, repoName)
    if err != nil {
        return "", fmt.Errorf("failed to get PR branch: %w", err)
    }
    return branch, nil
}

// selectBranchInteractive selects branch with fzf
func (opts *addOptions) selectBranchInteractive() (string, error) {
    if !opts.fzfSelector.IsAvailable() {
        return "", fmt.Errorf("fzf is not installed. Please install fzf for interactive selection, or specify a branch name")
    }
    
    branches, err := git.ListBranchesWithExecutor(opts.executor)
    if err != nil {
        return "", fmt.Errorf("failed to list branches: %w", err)
    }
    
    if len(branches) == 0 {
        return "", fmt.Errorf("no branches found")
    }
    
    selected, err := opts.fzfSelector.Select(fzf.SelectOptions{
        Items:   branches,
        Prompt:  "Select branch: ",
        Reverse: true,
    })
    
    return selected, err
}

// checkExistingWorktree checks for existing worktree
func (opts *addOptions) checkExistingWorktree(branch string) error {
    existing, err := git.FindWorktreeWithExecutor(opts.executor, branch)
    if err != nil {
        return fmt.Errorf("failed to check existing worktree: %w", err)
    }
    if existing != nil {
        fmt.Printf("Worktree already exists: %s\n", existing.Path)
        return fmt.Errorf("worktree already exists")
    }
    return nil
}

// ensureBranchExists ensures branch exists and fetches if necessary
func (opts *addOptions) ensureBranchExists() error {
    // Skip if creating new branch
    if opts.createBranch && opts.prIdentifier == "" {
        return nil
    }
    
    exists, err := git.BranchExistsWithExecutor(opts.executor, opts.branch)
    if err != nil {
        return fmt.Errorf("failed to check branch: %w", err)
    }
    
    if !exists {
        return opts.fetchBranchIfRemoteExists()
    }
    
    return nil
}

// fetchBranchIfRemoteExists fetches if remote branch exists
func (opts *addOptions) fetchBranchIfRemoteExists() error {
    remoteExists, err := git.RemoteBranchExistsWithExecutor(opts.executor, opts.branch)
    if err != nil {
        return fmt.Errorf("failed to check remote branch: %w", err)
    }
    
    if remoteExists {
        fmt.Printf("Fetching branch %s from origin...\n", opts.branch)
        if err := git.FetchBranchWithExecutor(opts.executor, opts.branch); err != nil {
            return fmt.Errorf("failed to fetch branch: %w", err)
        }
        return nil
    }
    
    return fmt.Errorf("branch %s does not exist (use -b to create)", opts.branch)
}
```

**Migration Procedure**:
1. Rename current `runAdd` to `runAdd_old`
2. Implement new `runAdd` and helper functions
3. Verify existing tests pass with new implementation
4. Delete `runAdd_old`

## 4. File Change List

### New Files
- `internal/shell/executor.go`: Command execution abstraction interface and implementation
- `internal/shell/executor_test.go`: Executor unit tests (coverage target: 85%)
- `internal/fzf/selector.go`: fzf integration abstraction interface and implementation
- `internal/fzf/selector_test.go`: Selector unit tests (coverage target: 80%)

### Modified Files
- `cmd/add.go`: Split runAdd refactoring, DI of Executor/Selector
- `cmd/sw.go`: Change to use FzfSelector
- `cmd/rm.go`: Change to use FzfSelector
- `cmd/fzf.go`: Temporarily change to wrapper using FzfSelector (eventually delete)
- `internal/git/worktree.go`: DI Executor into all functions (keep existing functions as wrappers for compatibility)
- `internal/github/pr.go`: DI Executor into GetPRBranch

### Deleted Files
- None (No deletion in Phase 1. Plan to delete `cmd/fzf.go` in Phase 2)

## 5. Implementation Procedure

### Step 1: Implement Shell Abstraction Layer
**Files to Create**: 
- `internal/shell/executor.go`
- `internal/shell/executor_test.go`

**Implementation Content**:
1. Define `Executor` interface (Execute, ExecuteWithStdio, LookPath)
2. Implement `RealExecutor` (exec.Command wrapper)
3. Implement `MockExecutor` (for testing)
4. Create unit tests (normal and error cases)

**Testing**:
- Verify `RealExecutor` can execute basic commands like `echo`, `ls`
- Verify `MockExecutor` can return arbitrary values
- Test error handling

**Verification Commands**:
```bash
docker compose run --rm dev go test -v ./internal/shell/...
docker compose run --rm dev go test -cover ./internal/shell/...
```

### Step 2: DI Executor into git Package
**Files to Modify**: 
- `internal/git/worktree.go`

**Implementation Content**:
1. Add `WithExecutor` suffix versions for each function (e.g., `ListWithExecutor`)
2. Keep existing functions as thin wrappers using `RealExecutor`
3. Replace all `exec.Command` calls with `executor.Execute` / `ExecuteWithStdio`

**Testing**:
- Verify existing tests continue to pass
- Add new tests using `MockExecutor`

**Verification Commands**:
```bash
docker compose run --rm dev go test -v ./internal/git/...
docker compose run --rm dev go test -cover ./internal/git/...
```

### Step 3: Implement fzf Integration
**Files to Create**: 
- `internal/fzf/selector.go`
- `internal/fzf/selector_test.go`

**Implementation Content**:
1. Define `Selector` interface
2. Design `SelectOptions` structure
3. Implement `FzfSelector` (DI Executor)
4. Handle Ctrl+C (exit code 130)
5. Implement `MockSelector`

**Testing**:
- Simulate selection using `MockSelector`
- Test cancellation handling
- Test multiple selection

**Verification Commands**:
```bash
docker compose run --rm dev go test -v ./internal/fzf/...
docker compose run --rm dev go test -cover ./internal/fzf/...
```

### Step 4: Split runAdd in cmd/add.go
**Files to Modify**: 
- `cmd/add.go`

**Implementation Content**:
1. Introduce `addOptions` structure
2. Split `runAdd` main process into 6 steps
3. Extract each step into dedicated method:
   - `determineBranch`
   - `getBranchFromPR`
   - `selectBranchInteractive`
   - `checkExistingWorktree`
   - `ensureBranchExists`
   - `fetchBranchIfRemoteExists`
4. DI Executor and FzfSelector

**Testing**:
- Verify existing `cmd/add_test.go` continues to pass
- Add unit tests for new helper functions

**Verification Commands**:
```bash
docker compose run --rm dev go test -v ./cmd/ -run TestAdd
docker compose run --rm dev gofmt -l cmd/add.go  # Format check
```

### Step 5: Migrate cmd/sw.go and cmd/rm.go to FzfSelector
**Files to Modify**: 
- `cmd/sw.go`
- `cmd/rm.go`
- `cmd/fzf.go` (update as wrapper)

**Implementation Content**:
1. Change `selectWorktreeWithFzf` in `cmd/fzf.go` to use `fzf.Selector`
2. DI `fzf.Selector` in `cmd/sw.go` and `cmd/rm.go`
3. Unify error messages

**Testing**:
- Verify `cmd/sw_test.go` and `cmd/rm_test.go` continue to pass

**Verification Commands**:
```bash
docker compose run --rm dev go test -v ./cmd/ -run TestSw
docker compose run --rm dev go test -v ./cmd/ -run TestRm
```

### Step 6: DI Executor into github Package
**Files to Modify**: 
- `internal/github/pr.go`

**Implementation Content**:
1. Add `WithExecutor` version to `GetPRBranch`
2. Replace `gh` command calls with `executor.Execute`

**Testing**:
- Verify existing tests continue to pass
- Add new tests using `MockExecutor`

**Verification Commands**:
```bash
docker compose run --rm dev go test -v ./internal/github/...
docker compose run --rm dev go test -cover ./internal/github/...
```

### Step 7: Integration Tests and Documentation Update
**Implementation Content**:
1. Run tests for all packages
2. Generate test coverage report
3. Verify build
4. Simple functional verification (run gw command in Docker)

**Verification Commands**:
```bash
# Run all tests
docker compose run --rm dev go test -v ./...

# Coverage report
docker compose run --rm dev go test -coverprofile=coverage.out ./...
docker compose run --rm dev go tool cover -func=coverage.out

# Verify build
docker compose run --rm dev go build -o gw .

# Simple functional verification
docker compose run --rm dev ./gw --help
```

### Step 8: Code Quality Check
**Implementation Content**:
1. Format check with go fmt
2. Static analysis with go vet
3. Check cyclomatic complexity (gocyclo)

**Verification Commands**:
```bash
docker compose run --rm dev gofmt -l .
docker compose run --rm dev go vet ./...
docker compose run --rm dev sh -c "go install github.com/fzipp/gocyclo/cmd/gocyclo@latest && gocyclo -over 10 ."
```

### Step 9: Regression Testing
**Implementation Content**:
1. Verify all existing test cases continue to pass
2. Verify newly added test cases
3. Test edge cases (empty repository, without fzf, etc.)

**Verification Commands**:
```bash
docker compose run --rm dev go test -v -count=1 ./...
```

### Step 10: Phase 1 Completion Verification
**Implementation Content**:
1. Check completion criteria checklist
2. Review changes
3. Prepare for transition to Phase 2

**Verification Commands**:
```bash
# Coverage summary
docker compose run --rm dev go test -cover ./...

# Complexity check
docker compose run --rm dev sh -c "gocyclo -over 10 cmd/add.go"
```

## 6. Test Plan

### Test Coverage Targets
- `internal/shell`: **85%** (All Executor methods + error handling)
- `internal/fzf`: **80%** (Selector selection logic + cancellation handling)
- `internal/git`: **40%** (Improve from current 8.5%, add WithExecutor versions to all functions)
- `internal/github`: **50%** (Improve from current 14.6%)
- `cmd`: **30%** (Improve from current 7.9%, test main flows)

### Test Cases

#### internal/shell/executor_test.go
- `RealExecutor.Execute`: Normal command execution, error handling
- `RealExecutor.ExecuteWithStdio`: Verify standard I/O connection
- `RealExecutor.LookPath`: Verify command existence check
- `MockExecutor`: Verify mock behavior for each method

#### internal/fzf/selector_test.go
- `FzfSelector.Select`: Normal case for single selection
- `FzfSelector.SelectMulti`: Normal case for multiple selection
- `FzfSelector.IsAvailable`: Verify fzf existence check
- Cancellation handling (exit code 130)
- Empty selection result handling
- `MockSelector`: Verify mock behavior for each method

#### internal/git/worktree_test.go (additions)
- `ListWithExecutor`: Get worktree list using mock
- `AddWithExecutor`: Create worktree using mock
- `BranchExistsWithExecutor`: Check branch existence using mock

#### cmd/add_test.go (additions)
- `determineBranch`: Branch determination from arguments/PR/fzf
- `checkExistingWorktree`: Detect existing worktree
- `ensureBranchExists`: Check branch existence and fetch
- Various error handling scenarios

## 7. Risk Management

### Risk 1: Breaking Existing Code with Executor DI
- **Impact**: HIGH
- **Mitigation**: 
  - Keep existing functions as thin wrappers for backward compatibility
  - Add new functions with `WithExecutor` suffix for gradual migration
  - Verify existing tests pass at each step

### Risk 2: fzf Integration Implementation Becomes Complex
- **Impact**: MEDIUM
- **Mitigation**: 
  - Limit direct `exec.Command` usage to `selectInternal` only
  - Implement directly without using `Executor` due to stdin connection requirement
  - Clearly document error handling (especially Ctrl+C)

### Risk 3: Test Coverage Doesn't Reach Target
- **Impact**: MEDIUM
- **Mitigation**: 
  - Proceed with test-first approach when implementing new packages in Steps 1-3
  - Check coverage at each step and fill gaps
  - Use mocks to eliminate external dependencies

### Risk 4: Build/Test Failure in Docker Environment
- **Impact**: LOW
- **Mitigation**: 
  - Run tests with `docker compose run --rm dev` at each step
  - Verified that required dependencies (git, fzf) are included in Dockerfile
  - Go 1.23 compatibility confirmed

### Risk 5: Complexity Migration from runAdd Split
- **Impact**: LOW
- **Mitigation**: 
  - Design each helper function to have single responsibility
  - Clearly correspond function names with processing content
  - Measure complexity of each function individually, maintain below 10

## 8. Implementation Checklist

### Before Phase 1 Start
- [x] Verify Analyze agent analysis results
- [x] Verify current codebase
- [x] Verify Docker environment works
- [x] Create implementation plan

### Step 1: Shell Abstraction Layer
- [ ] Create `internal/shell/executor.go`
- [ ] Define `Executor` interface
- [ ] Implement `RealExecutor`
- [ ] Implement `MockExecutor`
- [ ] Create `internal/shell/executor_test.go`
- [ ] Achieve test coverage 85% or above
- [ ] Run tests: `go test -v ./internal/shell/...`

### Step 2: git Package DI
- [ ] Add `WithExecutor` functions to `internal/git/worktree.go`
- [ ] Replace all `exec.Command` calls
- [ ] Keep existing functions as wrappers
- [ ] Verify existing tests pass
- [ ] Add new tests

### Step 3: fzf Integration
- [ ] Create `internal/fzf/selector.go`
- [ ] Define `Selector` interface
- [ ] Implement `FzfSelector`
- [ ] Implement Ctrl+C handling
- [ ] Implement `MockSelector`
- [ ] Create `internal/fzf/selector_test.go`
- [ ] Achieve test coverage 80% or above
- [ ] Run tests: `go test -v ./internal/fzf/...`

### Step 4: Split runAdd
- [ ] Add `addOptions` structure to `cmd/add.go`
- [ ] Implement `determineBranch`
- [ ] Implement `getBranchFromPR`
- [ ] Implement `selectBranchInteractive`
- [ ] Implement `checkExistingWorktree`
- [ ] Implement `ensureBranchExists`
- [ ] Implement `fetchBranchIfRemoteExists`
- [ ] Implement new `runAdd`
- [ ] Verify complexity is 10 or below
- [ ] Verify existing tests pass

### Step 5: Migrate sw/rm to FzfSelector
- [ ] Change `cmd/fzf.go` to use `fzf.Selector`
- [ ] Modify `cmd/sw.go`
- [ ] Modify `cmd/rm.go`
- [ ] Verify existing tests pass

### Step 6: github Package DI
- [ ] Add `WithExecutor` function to `internal/github/pr.go`
- [ ] Verify existing tests pass
- [ ] Add new tests

### Step 7: Integration Tests
- [ ] Run tests for all packages
- [ ] Generate coverage report
- [ ] Verify build success
- [ ] Functional verification in Docker

### Step 8: Code Quality
- [ ] Format check with `go fmt`
- [ ] Static analysis with `go vet`
- [ ] Complexity check with `gocyclo`

### Step 9: Regression Testing
- [ ] All existing tests pass
- [ ] Edge case testing

### Phase 1 Completion Verification
- [ ] Shell abstraction layer complete (15 `exec.Command` migrations complete)
- [ ] fzf integration complete (3 locations unified)
- [ ] `runAdd` complexity below 10 achieved
- [ ] Test coverage targets achieved:
  - [ ] `internal/shell`: 85%+
  - [ ] `internal/fzf`: 80%+
  - [ ] `internal/git`: 40%+
  - [ ] `internal/github`: 50%+
  - [ ] `cmd`: 30%+
- [ ] All tests pass
- [ ] Build success in Docker environment
- [ ] Ready for Phase 2 transition

---

## Additional Information

### Utilizing Go 1.23
This project uses Go 1.23, so the following features can be utilized:
- Iterator pattern (range over function)
- Enhanced `slices` package
- Convenient functions like `cmp.Or`

However, Phase 1 focuses on basic refactoring, so introduction of these advanced features will be considered in Phase 2 and beyond.

### Integration with Cobra Framework
- The `cmd/*` structure follows Cobra conventions
- Use functions that return errors in the `RunE` field
- Define flags in `init()`

### Docker Development Environment
```bash
# Run tests
docker compose run --rm dev go test -v ./...

# Build
docker compose run --rm dev go build -o gw .

# Coverage
docker compose run --rm dev go test -coverprofile=coverage.out ./...
docker compose run --rm dev go tool cover -html=coverage.out -o coverage.html
```

### Implementation Notes
1. **Backward Compatibility**: Don't delete existing functions, add new `WithExecutor` versions
2. **Test First**: Implement new packages from interface and tests
3. **Gradual Migration**: Verify functionality at each step, address issues immediately
4. **Documentation**: Write appropriate comments for each interface and structure
