# gw - Git Worktree Wrapper

Git worktree をシンプルに管理するための CLI ツール。

## 特徴

- 📁 直感的なワークツリー作成（`gw add feature/hoge` → `../repo-feature-hoge/`）
- 🔀 ブランチ名、サフィックス、ディレクトリ名の柔軟な指定
- 🐙 GitHub PR からのワークツリー作成
- 🔍 fzf によるインタラクティブなワークツリー選択
- 🚀 シェル統合によるスムーズなディレクトリ移動

## インストール

### Homebrew (macOS/Linux)

```bash
brew install t98o84/tap/gw
```

### Go

```bash
go install github.com/t98o84/gw@latest
```

### バイナリ

[Releases](https://github.com/t98o84/gw/releases) からダウンロード。

### ソースからビルド

```bash
# リポジトリをクローン
git clone https://github.com/t98o84/gw.git
cd gw

# Docker でビルド (macOS Apple Silicon)
docker compose run --rm dev sh -c "CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o gw ."

# Docker でビルド (macOS Intel)
docker compose run --rm dev sh -c "CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o gw ."

# Docker でビルド (Linux)
docker compose run --rm dev go build -o gw .

# パスの通った場所にコピー
sudo cp gw /usr/local/bin/
# または
mkdir -p ~/.local/bin && cp gw ~/.local/bin/
```

ローカルに Go がインストールされている場合：

```bash
go install github.com/t98o84/gw@latest
```

## シェル統合のセットアップ

`gw sw` でディレクトリ移動するために、シェル設定に以下を追加してください：

### Bash

```bash
# ~/.bashrc に追加
eval "$(gw init bash)"
```

### Zsh

```bash
# ~/.zshrc に追加
eval "$(gw init zsh)"
```

### Fish

```fish
# ~/.config/fish/config.fish に追加
gw init fish | source
```

## 使い方

### 設定ファイル

`gw` は YAML 形式の設定ファイルをサポートしています。設定ファイルのパスは以下の通りです：

- **Linux/macOS**: `~/.config/gw/config.yaml` (または `$XDG_CONFIG_HOME/gw/config.yaml`)
- **Windows**: `%APPDATA%\gw\config.yaml`

#### 設定例

```yaml
add:
  open: true  # ワークツリー作成後に自動的にエディターで開く
rm:
  branch: false  # ワークツリー削除時にブランチも削除する
editor: code  # 使用するエディターコマンド
```

#### 設定項目

- `add.open` (boolean): ワークツリー作成後に自動的にエディターで開くかどうか（デフォルト: `false`）
- `rm.branch` (boolean): ワークツリー削除時に関連するブランチも削除するかどうか（デフォルト: `false`）
- `editor` (string): 使用するエディターコマンド（例: `code`, `vim`, `emacs`）

**注意**: コマンドラインフラグは設定ファイルの値より優先されます。

### ワークツリーの作成

```bash
# 既存ブランチのワークツリーを作成
gw add feature/hoge
# => ../ex-repo-feature-hoge/ が作成される

# 新規ブランチを作成してワークツリーを作成
gw add -b feature/new

# PR のブランチからワークツリーを作成
gw add --pr 123
gw add -p 123
gw add --pr https://github.com/owner/repo/pull/123
gw add -p https://github.com/owner/repo/pull/123

# ワークツリー作成後にエディターで開く（コマンドラインフラグ）
gw add --open --editor code feature/hoge
gw add --open -e vim feature/hoge

# 設定ファイルで add.open=true と editor=code を設定している場合
# フラグなしでもエディターが自動的に開く
gw add feature/hoge

# オプションの組み合わせも可能
gw add -b --open --editor code feature/new
gw add --pr 123 --open -e vim
```

### ワークツリー一覧

```bash
gw ls
# ex-repo (main)
# ex-repo-feature-hoge
# ex-repo-fix-foo
```

### ワークツリーの削除

```bash
# 以下はすべて同じワークツリーを指定
gw rm feature/hoge
gw rm feature-hoge
gw rm ex-repo-feature-hoge

# ブランチも一緒に削除（-b/--branch オプション）
gw rm -b feature/hoge
gw rm --branch feature-hoge

# 強制削除（マージされていないブランチも削除）
gw rm -f -b feature/hoge
```

**注意**: ブランチ削除には以下の安全性チェックが適用されます：
- `main` または `master` ブランチは削除できません
- カレントブランチは削除できません
- マージされていないブランチは `-f`/`--force` フラグなしでは削除できません

### ワークツリーでコマンド実行

```bash
gw exec feature/hoge git status
gw exec feature-hoge npm install
```

### ワークツリーへ移動

```bash
# 指定したワークツリーに移動
gw sw feature/hoge

# fzf でインタラクティブに選択
gw sw
```

## コマンド一覧

| コマンド | エイリアス | 説明 |
|---------|-----------|------|
| `gw add <branch>` | `gw a` | ワークツリー作成 |
| `gw add -b <branch>` | `gw a -b` | 新規ブランチ + ワークツリー作成 |
| `gw add --pr <url\|number>` | `gw a --pr`, `gw a -p` | PR ブランチのワークツリー作成 |
| `gw add --open` | `gw a --open` | ワークツリー作成後にエディターで開く |
| `gw add --editor <cmd>` | `gw a -e` | 使用するエディターコマンドを指定 |
| `gw ls` | `gw l` | ワークツリー一覧表示 |
| `gw rm <name>` | `gw r` | ワークツリー削除 |
| `gw rm -b <name>` | `gw r -b` | ワークツリーとブランチを削除 |
| `gw exec <name> <cmd...>` | `gw e` | 対象ワークツリーでコマンド実行 |
| `gw sw [name]` | `gw s` | 対象ワークツリーに移動（引数なしで fzf） |
| `gw fd` | `gw f` | fzf でワークツリー検索 |
| `gw init <shell>` | `gw i` | シェル初期化スクリプト出力 |

## 必要なツール

- `git`
- `fzf` (オプション: インタラクティブ選択用)
- `gh` または `GITHUB_TOKEN` 環境変数 (オプション: PR 連携用)

## ライセンス

MIT
