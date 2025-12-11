# AGENTS.md - AI エージェント向け開発ガイド

このドキュメントは AI エージェント（GitHub Copilot、Claude 等）がこのプロジェクトで作業する際のガイドラインです。

## 開発環境

### ⚠️ 重要: Docker 環境を使用すること

**ローカル環境ではなく、必ず Docker 環境で開発・ビルド・テストを行ってください。**

```bash
# 開発コンテナに入る
docker compose run --rm dev sh

# または直接コマンドを実行
docker compose run --rm dev go test ./...
docker compose run --rm dev go build -o gw .
```

### コマンド例

```bash
# テスト実行
docker compose run --rm dev go test ./...

# 詳細なテスト出力
docker compose run --rm dev go test ./... -v

# ビルド（Linux 向け）
docker compose run --rm dev go build -o gw .

# ビルド（macOS Apple Silicon 向け）
docker compose run --rm dev sh -c "CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o gw ."

# ビルド（macOS Intel 向け）
docker compose run --rm dev sh -c "CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o gw ."

# go mod tidy
docker compose run --rm dev go mod tidy

# フォーマット
docker compose run --rm dev go fmt ./...
```

## プロジェクト構成

```
gw/
├── main.go              # エントリーポイント
├── cmd/                 # Cobra コマンド
│   ├── root.go          # ルートコマンド
│   ├── add.go           # gw add - ワークツリー作成
│   ├── rm.go            # gw rm - ワークツリー削除（複数選択対応）
│   ├── ls.go            # gw ls - ワークツリー一覧
│   ├── sw.go            # gw sw - ワークツリー切り替え
│   ├── exec.go          # gw exec - ワークツリーでコマンド実行
│   ├── fd.go            # gw fd - fzf でワークツリー検索
│   ├── init.go          # gw init - シェル統合スクリプト出力
│   └── fzf.go           # fzf ヘルパー関数
├── internal/
│   ├── git/             # Git 操作
│   │   ├── worktree.go  # git worktree 操作
│   │   └── naming.go    # 命名規則変換
│   └── github/          # GitHub API
│       └── pr.go        # PR からブランチ取得
├── go.mod
├── go.sum
├── Dockerfile
└── compose.yaml
```

## コーディング規約

### 言語
- コード内コメント: 英語
- コミットメッセージ: 英語
- ドキュメント（README 等）: 日本語

### スタイル
- Go の標準的なフォーマット（`go fmt`）に従う
- エラーは適切にラップして返す（`fmt.Errorf("context: %w", err)`）
- 外部コマンド実行時は `os/exec` を使用

### テスト
- テストファイルは `*_test.go` の命名規則
- fzf など対話的な入力が必要な関数は直接呼び出さない
- テーブル駆動テストを推奨

## 依存関係

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI フレームワーク
- [github.com/google/go-github](https://github.com/google/go-github) - GitHub API クライアント
- [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2) - OAuth2 認証

## 注意事項

1. **fzf 関連のテスト**: fzf は対話的な入力を必要とするため、テストでは直接呼び出さないこと
2. **git コマンド**: `internal/git` パッケージ経由で実行。テスト時は git リポジトリ外で実行される可能性がある
3. **GitHub API**: 認証には `GITHUB_TOKEN`、`GH_TOKEN` 環境変数、または `gh auth token` を使用
