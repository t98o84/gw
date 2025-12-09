# Phase 1 実装計画書

## 1. 実装スコープ

### 対応項目
- **shell 抽象化レイヤーの導入**: `exec.Command` の直接呼び出しを抽象化し、テスト可能にする
- **fzf 統合の共通化**: 3箇所に分散している fzf 呼び出しを統一インターフェースに集約
- **runAdd 関数の分割**: 96行の複雑な関数を責務ごとに分割し、保守性を向上

### 非対応項目 (Phase 2以降)
- グローバル変数のフラグ管理の改善 (Phase 2)
- エラーハンドリングの統一的なラッパー実装 (Phase 2)
- GitHub API との統合テスト (Phase 3)
- パフォーマンス最適化 (Phase 3)

### 完了基準
- [ ] `internal/shell/executor.go` が実装され、全 15箇所の `exec.Command` 呼び出しが移行完了
- [ ] `internal/fzf/selector.go` が実装され、3箇所の fzf 呼び出しが統一インターフェースに移行
- [ ] `cmd/add.go:runAdd` の複雑度が 10 以下に低減
- [ ] 新規パッケージ (`internal/shell`, `internal/fzf`) のテストカバレッジが 80% 以上
- [ ] 既存の全テストが引き続きパス
- [ ] Docker 環境でビルドとテストが正常に実行できる

## 2. アーキテクチャ設計

### パッケージ構造

```
gw/
├── cmd/                        # CLI コマンド層
│   ├── add.go                 # リファクタリング: Executor/FzfSelector 使用
│   ├── sw.go                  # リファクタリング: FzfSelector 使用
│   ├── rm.go                  # リファクタリング: FzfSelector 使用
│   ├── fzf.go                 # 削除候補: fzf パッケージに統合
│   └── ...
├── internal/
│   ├── shell/                 # 新規: コマンド実行抽象化
│   │   ├── executor.go        # interface + 実装
│   │   └── executor_test.go   # モック含むテスト
│   ├── fzf/                   # 新規: fzf 統合
│   │   ├── selector.go        # interface + 実装
│   │   └── selector_test.go   # モック含むテスト
│   ├── git/
│   │   ├── worktree.go        # リファクタリング: Executor 使用
│   │   ├── naming.go          # 変更なし
│   │   └── ...
│   └── github/
│       └── pr.go              # リファクタリング: Executor 使用
```

### 依存関係

```
cmd/* 
  ↓ (使用)
internal/shell/Executor ←┐
internal/fzf/Selector    │
  ↓ (使用)               │
internal/git/*           │
internal/github/*        │
  ↓ (使用)               │
internal/shell/Executor ─┘
```

**設計原則**:
- `cmd` パッケージは `internal/shell` と `internal/fzf` の interface のみに依存
- `internal/git` と `internal/github` は `internal/shell.Executor` を DI で受け取る
- テスト時は `MockExecutor` を注入してコマンド実行をシミュレート

## 3. 詳細設計

### 3.1. shell 抽象化レイヤー

**ファイル**: `internal/shell/executor.go`

```go
package shell

import (
    "io"
    "os/exec"
)

// Executor はシェルコマンドの実行を抽象化するインターフェース
type Executor interface {
    // Execute はコマンドを実行し、標準出力を返す
    Execute(name string, args ...string) ([]byte, error)
    
    // ExecuteWithStdio はコマンドを実行し、標準入出力を接続する
    ExecuteWithStdio(name string, args ...string) error
    
    // LookPath はコマンドが存在するか確認する
    LookPath(name string) (string, error)
}

// RealExecutor は実際のコマンド実行を行う実装
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

// MockExecutor はテスト用のモック実装
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

**移行方針**:
1. `internal/git/worktree.go` の全関数に `Executor` を DI
2. `internal/github/pr.go` の `GetPRBranch` に `Executor` を DI
3. `cmd/*` で `RealExecutor` を初期化して渡す

### 3.2. fzf 統合の共通化

**ファイル**: `internal/fzf/selector.go`

```go
package fzf

import (
    "fmt"
    "os"
    "strings"
    
    "github.com/t98o84/gw/internal/shell"
)

// SelectOptions は fzf の選択オプション
type SelectOptions struct {
    Items       []string       // 選択肢のリスト
    Prompt      string         // プロンプトメッセージ
    Multi       bool           // 複数選択を許可
    Height      string         // 表示高さ (default: "40%")
    Reverse     bool           // 逆順表示
}

// Selector は fzf による対話的選択を抽象化するインターフェース
type Selector interface {
    // Select は単一選択を行う
    Select(opts SelectOptions) (string, error)
    
    // SelectMulti は複数選択を行う
    SelectMulti(opts SelectOptions) ([]string, error)
    
    // IsAvailable は fzf が利用可能か確認する
    IsAvailable() bool
}

// FzfSelector は fzf コマンドを使用した実装
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
    // fzf 引数の構築
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
    
    // fzf 実行 (標準入力に items を渡す必要があるため、特殊処理)
    cmd := exec.Command("fzf", args...)
    cmd.Stdin = strings.NewReader(strings.Join(opts.Items, "\n"))
    cmd.Stderr = os.Stderr
    
    out, err := cmd.Output()
    if err != nil {
        // Ctrl+C (exit code 130) はエラーではなくキャンセル扱い
        if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
            return nil, nil
        }
        return nil, nil // その他のエラーもキャンセル扱い
    }
    
    output := strings.TrimSpace(string(out))
    if output == "" {
        return nil, nil
    }
    
    // 複数行の結果を分割
    results := strings.Split(output, "\n")
    var trimmed []string
    for _, r := range results {
        if t := strings.TrimSpace(r); t != "" {
            trimmed = append(trimmed, t)
        }
    }
    
    return trimmed, nil
}

// MockSelector はテスト用のモック実装
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

**移行方針**:
1. `cmd/fzf.go` の `selectWorktreeWithFzf` を `fzf.Selector` 使用に書き換え
2. `cmd/add.go` の `selectBranchWithFzf` を `fzf.Selector` 使用に書き換え
3. `cmd/fzf.go` は最終的に削除し、機能を `internal/fzf` に完全移行

### 3.3. runAdd 関数の分割

**分割後の構成**:

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
    
    // Step 1: ブランチ名の決定
    branch, err := opts.determineBranch(args)
    if err != nil {
        return err
    }
    if branch == "" {
        return nil // User cancelled
    }
    opts.branch = branch
    
    // Step 2: リポジトリ情報の取得
    repoName, err := git.GetRepoName()
    if err != nil {
        return fmt.Errorf("failed to get repository name: %w", err)
    }
    
    // Step 3: Worktree パスの生成
    wtPath, err := git.WorktreePath(repoName, branch)
    if err != nil {
        return fmt.Errorf("failed to generate worktree path: %w", err)
    }
    
    // Step 4: 既存 worktree のチェック
    if err := opts.checkExistingWorktree(branch); err != nil {
        return err
    }
    
    // Step 5: ブランチの存在確認と fetch
    if err := opts.ensureBranchExists(); err != nil {
        return err
    }
    
    // Step 6: Worktree の作成
    fmt.Printf("Creating worktree at %s for branch %s...\n", wtPath, branch)
    if err := git.AddWithExecutor(executor, wtPath, branch, opts.createBranch); err != nil {
        return err
    }
    
    fmt.Printf("✓ Worktree created: %s\n", wtPath)
    return nil
}

// determineBranch はブランチ名を決定する (PR/fzf/引数から)
func (opts *addOptions) determineBranch(args []string) (string, error) {
    // PR 指定の場合
    if opts.prIdentifier != "" {
        return opts.getBranchFromPR()
    }
    
    // 引数なしの場合は fzf で選択
    if len(args) == 0 {
        return opts.selectBranchInteractive()
    }
    
    // 引数で指定された場合
    return args[0], nil
}

// getBranchFromPR は PR からブランチ名を取得
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

// selectBranchInteractive は fzf でブランチを選択
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

// checkExistingWorktree は既存の worktree をチェック
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

// ensureBranchExists はブランチの存在を確認し、必要に応じて fetch
func (opts *addOptions) ensureBranchExists() error {
    // 新規ブランチ作成の場合はスキップ
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

// fetchBranchIfRemoteExists はリモートブランチが存在すれば fetch
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

**移行手順**:
1. 現在の `runAdd` を `runAdd_old` にリネーム
2. 新しい `runAdd` と補助関数を実装
3. 既存テストが新実装でもパスすることを確認
4. `runAdd_old` を削除

## 4. ファイル変更一覧

### 新規作成
- `internal/shell/executor.go`: コマンド実行の抽象化 interface と実装
- `internal/shell/executor_test.go`: Executor のユニットテスト (カバレッジ目標: 85%)
- `internal/fzf/selector.go`: fzf 統合の抽象化 interface と実装
- `internal/fzf/selector_test.go`: Selector のユニットテスト (カバレッジ目標: 80%)

### 変更
- `cmd/add.go`: runAdd の分割リファクタリング、Executor/Selector の DI
- `cmd/sw.go`: FzfSelector の使用に変更
- `cmd/rm.go`: FzfSelector の使用に変更
- `cmd/fzf.go`: 一時的に FzfSelector を使用するラッパーに変更 (最終的に削除)
- `internal/git/worktree.go`: 全関数に Executor を DI (既存関数は互換性のためラッパーとして残す)
- `internal/github/pr.go`: GetPRBranch に Executor を DI

### 削除
- なし (Phase 1 では削除なし。Phase 2 で `cmd/fzf.go` を削除予定)

## 5. 実装手順

### Step 1: shell 抽象化レイヤーの実装
**作成するファイル**: 
- `internal/shell/executor.go`
- `internal/shell/executor_test.go`

**実装内容**:
1. `Executor` interface の定義 (Execute, ExecuteWithStdio, LookPath)
2. `RealExecutor` の実装 (exec.Command のラッパー)
3. `MockExecutor` の実装 (テスト用)
4. ユニットテストの作成 (正常系、エラー系)

**テスト**:
- `RealExecutor` が `echo`, `ls` などの基本コマンドを実行できることを確認
- `MockExecutor` が任意の戻り値を返せることを確認
- エラーハンドリングのテスト

**確認コマンド**:
```bash
docker compose run --rm dev go test -v ./internal/shell/...
docker compose run --rm dev go test -cover ./internal/shell/...
```

### Step 2: git パッケージへの Executor DI
**変更するファイル**: 
- `internal/git/worktree.go`

**実装内容**:
1. 各関数に `WithExecutor` サフィックス版を追加 (例: `ListWithExecutor`)
2. 既存関数は `RealExecutor` を使う薄いラッパーとして維持
3. `exec.Command` の呼び出しを全て `executor.Execute` / `ExecuteWithStdio` に置き換え

**テスト**:
- 既存テストが引き続きパスすることを確認
- `MockExecutor` を使った新規テストを追加

**確認コマンド**:
```bash
docker compose run --rm dev go test -v ./internal/git/...
docker compose run --rm dev go test -cover ./internal/git/...
```

### Step 3: fzf 統合の実装
**作成するファイル**: 
- `internal/fzf/selector.go`
- `internal/fzf/selector_test.go`

**実装内容**:
1. `Selector` interface の定義
2. `SelectOptions` 構造体の設計
3. `FzfSelector` の実装 (Executor を DI)
4. Ctrl+C ハンドリング (exit code 130)
5. `MockSelector` の実装

**テスト**:
- `MockSelector` を使った選択シミュレーション
- キャンセル処理のテスト
- 複数選択のテスト

**確認コマンド**:
```bash
docker compose run --rm dev go test -v ./internal/fzf/...
docker compose run --rm dev go test -cover ./internal/fzf/...
```

### Step 4: cmd/add.go の runAdd 分割
**変更するファイル**: 
- `cmd/add.go`

**実装内容**:
1. `addOptions` 構造体の導入
2. `runAdd` のメイン処理を6ステップに分割
3. 各ステップを専用メソッドに抽出:
   - `determineBranch`
   - `getBranchFromPR`
   - `selectBranchInteractive`
   - `checkExistingWorktree`
   - `ensureBranchExists`
   - `fetchBranchIfRemoteExists`
4. Executor と FzfSelector を DI

**テスト**:
- 既存の `cmd/add_test.go` が引き続きパスすることを確認
- 新しい補助関数のユニットテストを追加

**確認コマンド**:
```bash
docker compose run --rm dev go test -v ./cmd/ -run TestAdd
docker compose run --rm dev gofmt -l cmd/add.go  # フォーマット確認
```

### Step 5: cmd/sw.go と cmd/rm.go の FzfSelector 移行
**変更するファイル**: 
- `cmd/sw.go`
- `cmd/rm.go`
- `cmd/fzf.go` (ラッパーとして更新)

**実装内容**:
1. `cmd/fzf.go` の `selectWorktreeWithFzf` を `fzf.Selector` を使う実装に変更
2. `cmd/sw.go` と `cmd/rm.go` で `fzf.Selector` を DI
3. エラーメッセージの統一

**テスト**:
- `cmd/sw_test.go` と `cmd/rm_test.go` が引き続きパスすることを確認

**確認コマンド**:
```bash
docker compose run --rm dev go test -v ./cmd/ -run TestSw
docker compose run --rm dev go test -v ./cmd/ -run TestRm
```

### Step 6: github パッケージへの Executor DI
**変更するファイル**: 
- `internal/github/pr.go`

**実装内容**:
1. `GetPRBranch` に `WithExecutor` 版を追加
2. `gh` コマンドの呼び出しを `executor.Execute` に置き換え

**テスト**:
- 既存テストが引き続きパスすることを確認
- `MockExecutor` を使った新規テストを追加

**確認コマンド**:
```bash
docker compose run --rm dev go test -v ./internal/github/...
docker compose run --rm dev go test -cover ./internal/github/...
```

### Step 7: 統合テストとドキュメント更新
**実施内容**:
1. 全パッケージのテストを実行
2. テストカバレッジレポートの生成
3. ビルド確認
4. 簡単な動作確認 (Docker 内で gw コマンドを実行)

**確認コマンド**:
```bash
# 全テスト実行
docker compose run --rm dev go test -v ./...

# カバレッジレポート
docker compose run --rm dev go test -coverprofile=coverage.out ./...
docker compose run --rm dev go tool cover -func=coverage.out

# ビルド確認
docker compose run --rm dev go build -o gw .

# 動作確認 (簡易)
docker compose run --rm dev ./gw --help
```

### Step 8: コード品質チェック
**実施内容**:
1. go fmt でフォーマット確認
2. go vet で静的解析
3. 循環的複雑度の確認 (gocyclo)

**確認コマンド**:
```bash
docker compose run --rm dev gofmt -l .
docker compose run --rm dev go vet ./...
docker compose run --rm dev sh -c "go install github.com/fzipp/gocyclo/cmd/gocyclo@latest && gocyclo -over 10 ."
```

### Step 9: リグレッションテスト
**実施内容**:
1. 既存の全テストケースが引き続きパスすることを確認
2. 新規追加したテストケースの確認
3. エッジケースのテスト (空のリポジトリ、fzf なしなど)

**確認コマンド**:
```bash
docker compose run --rm dev go test -v -count=1 ./...
```

### Step 10: Phase 1 完了確認
**実施内容**:
1. 完了基準のチェックリスト確認
2. 変更内容のレビュー
3. Phase 2 への移行準備

**確認コマンド**:
```bash
# カバレッジサマリー
docker compose run --rm dev go test -cover ./...

# 複雑度確認
docker compose run --rm dev sh -c "gocyclo -over 10 cmd/add.go"
```

## 6. テスト計画

### テストカバレッジ目標
- `internal/shell`: **85%** (Executor の全メソッド + エラーハンドリング)
- `internal/fzf`: **80%** (Selector の選択ロジック + キャンセル処理)
- `internal/git`: **40%** (現状 8.5% から向上、全関数に WithExecutor 版を追加)
- `internal/github`: **50%** (現状 14.6% から向上)
- `cmd`: **30%** (現状 7.9% から向上、主要フローのテスト)

### テストケース

#### internal/shell/executor_test.go
- `RealExecutor.Execute`: コマンド正常実行、エラー時の処理
- `RealExecutor.ExecuteWithStdio`: 標準入出力の接続確認
- `RealExecutor.LookPath`: コマンドの存在確認
- `MockExecutor`: 各メソッドのモック動作確認

#### internal/fzf/selector_test.go
- `FzfSelector.Select`: 単一選択の正常系
- `FzfSelector.SelectMulti`: 複数選択の正常系
- `FzfSelector.IsAvailable`: fzf の存在確認
- キャンセル処理 (exit code 130)
- 空の選択結果の処理
- `MockSelector`: 各メソッドのモック動作確認

#### internal/git/worktree_test.go (追加分)
- `ListWithExecutor`: モックを使った worktree リスト取得
- `AddWithExecutor`: モックを使った worktree 作成
- `BranchExistsWithExecutor`: モックを使ったブランチ存在確認

#### cmd/add_test.go (追加分)
- `determineBranch`: 引数/PR/fzf からのブランチ決定
- `checkExistingWorktree`: 既存 worktree の検出
- `ensureBranchExists`: ブランチ存在確認と fetch
- エラーハンドリング各種

## 7. リスク管理

### リスク1: Executor DI による既存コードの破壊
- **影響度**: HIGH
- **対策**: 
  - 既存関数を薄いラッパーとして残し、後方互換性を維持
  - `WithExecutor` サフィックスで新関数を追加し、段階的に移行
  - 各ステップで既存テストがパスすることを確認

### リスク2: fzf 統合の実装が複雑化
- **影響度**: MEDIUM
- **対策**: 
  - `exec.Command` を直接使用する部分は `selectInternal` に限定
  - 標準入力の接続が必要なため、`Executor` を使わず直接実装
  - エラーハンドリング (特に Ctrl+C) を明確に文書化

### リスク3: テストカバレッジが目標に達しない
- **影響度**: MEDIUM
- **対策**: 
  - Step 1-3 で新規パッケージを実装する際、テストファーストで進める
  - 各ステップでカバレッジを確認し、不足分を補う
  - モックを活用して外部依存を排除

### リスク4: Docker 環境でのビルド/テストの失敗
- **影響度**: LOW
- **対策**: 
  - 各ステップで `docker compose run --rm dev` でテスト実行
  - Dockerfile に必要な依存関係 (git, fzf) が含まれていることを確認済み
  - Go 1.23 の互換性は問題なし

### リスク5: runAdd 分割による複雑度の移動
- **影響度**: LOW
- **対策**: 
  - 各補助関数は単一責務を持つように設計
  - 関数名と処理内容を明確に対応させる
  - 各関数の複雑度を個別に測定し、10 以下を維持

## 8. 実装チェックリスト

### Phase 1 開始前
- [x] Analyze エージェントの分析結果を確認
- [x] 現在のコードベースを確認
- [x] Docker 環境が動作することを確認
- [x] 実装計画書を作成

### Step 1: shell 抽象化レイヤー
- [ ] `internal/shell/executor.go` 作成
- [ ] `Executor` interface 定義
- [ ] `RealExecutor` 実装
- [ ] `MockExecutor` 実装
- [ ] `internal/shell/executor_test.go` 作成
- [ ] テストカバレッジ 85% 以上達成
- [ ] テスト実行: `go test -v ./internal/shell/...`

### Step 2: git パッケージ DI
- [ ] `internal/git/worktree.go` に `WithExecutor` 関数追加
- [ ] 全 `exec.Command` 呼び出しを置き換え
- [ ] 既存関数をラッパーとして維持
- [ ] 既存テストがパスすることを確認
- [ ] 新規テスト追加

### Step 3: fzf 統合
- [ ] `internal/fzf/selector.go` 作成
- [ ] `Selector` interface 定義
- [ ] `FzfSelector` 実装
- [ ] Ctrl+C ハンドリング実装
- [ ] `MockSelector` 実装
- [ ] `internal/fzf/selector_test.go` 作成
- [ ] テストカバレッジ 80% 以上達成
- [ ] テスト実行: `go test -v ./internal/fzf/...`

### Step 4: runAdd 分割
- [ ] `cmd/add.go` に `addOptions` 構造体追加
- [ ] `determineBranch` 実装
- [ ] `getBranchFromPR` 実装
- [ ] `selectBranchInteractive` 実装
- [ ] `checkExistingWorktree` 実装
- [ ] `ensureBranchExists` 実装
- [ ] `fetchBranchIfRemoteExists` 実装
- [ ] 新 `runAdd` 実装
- [ ] 複雑度が 10 以下になったことを確認
- [ ] 既存テストがパスすることを確認

### Step 5: sw/rm の FzfSelector 移行
- [ ] `cmd/fzf.go` を `fzf.Selector` 使用に変更
- [ ] `cmd/sw.go` の変更
- [ ] `cmd/rm.go` の変更
- [ ] 既存テストがパスすることを確認

### Step 6: github パッケージ DI
- [ ] `internal/github/pr.go` に `WithExecutor` 関数追加
- [ ] 既存テストがパスすることを確認
- [ ] 新規テスト追加

### Step 7: 統合テスト
- [ ] 全パッケージのテスト実行
- [ ] カバレッジレポート生成
- [ ] ビルド成功確認
- [ ] Docker 内で動作確認

### Step 8: コード品質
- [ ] `go fmt` でフォーマット確認
- [ ] `go vet` で静的解析
- [ ] `gocyclo` で複雑度確認

### Step 9: リグレッションテスト
- [ ] 全既存テストがパス
- [ ] エッジケースのテスト

### Phase 1 完了確認
- [ ] shell 抽象化レイヤー完成 (15箇所の `exec.Command` 移行完了)
- [ ] fzf 統合完成 (3箇所の統一)
- [ ] `runAdd` 複雑度 10 以下達成
- [ ] テストカバレッジ目標達成:
  - [ ] `internal/shell`: 85%+
  - [ ] `internal/fzf`: 80%+
  - [ ] `internal/git`: 40%+
  - [ ] `internal/github`: 50%+
  - [ ] `cmd`: 30%+
- [ ] 全テストがパス
- [ ] Docker 環境でビルド成功
- [ ] Phase 2 移行準備完了

---

## 補足情報

### Go 1.23 の活用
このプロジェクトは Go 1.23 を使用しているため、以下の機能を活用できます:
- イテレータパターン (range over function)
- `slices` パッケージの強化
- `cmp.Or` などの便利関数

ただし、Phase 1 では基本的なリファクタリングに集中するため、これらの高度な機能の導入は Phase 2 以降で検討します。

### Cobra フレームワークとの統合
- `cmd/*` の構造は Cobra の規約に従っています
- `RunE` フィールドでエラーを返す関数を使用
- フラグは `init()` で定義

### Docker 開発環境
```bash
# テスト実行
docker compose run --rm dev go test -v ./...

# ビルド
docker compose run --rm dev go build -o gw .

# カバレッジ
docker compose run --rm dev go test -coverprofile=coverage.out ./...
docker compose run --rm dev go tool cover -html=coverage.out -o coverage.html
```

### 実装時の注意点
1. **後方互換性**: 既存の関数は削除せず、新しい `WithExecutor` 版を追加
2. **テストファースト**: 新規パッケージは interface とテストから実装
3. **段階的移行**: 各ステップで動作確認を行い、問題があれば即座に対処
4. **ドキュメント**: 各 interface と構造体には適切なコメントを記載
