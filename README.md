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
  sync: false  # メインワークツリーからファイルを同期する
  sync_ignored: false  # gitignored ファイルも同期する
rm:
  branch: false  # ワークツリー削除時にブランチも削除する
  force: false  # 確認プロンプトをスキップする
close:
  force: false  # 確認プロンプトをスキップする
editor: code  # 使用するエディターコマンド
```

#### 設定項目

- `add.open` (boolean): ワークツリー作成後に自動的にエディターで開くかどうか（デフォルト: `false`）
- `add.sync` (boolean): メインワークツリーからファイルを同期するかどうか（デフォルト: `false`）
- `add.sync_ignored` (boolean): gitignored ファイルも同期するかどうか（デフォルト: `false`）
- `rm.branch` (boolean): ワークツリー削除時に関連するブランチも削除するかどうか（デフォルト: `false`）
- `rm.force` (boolean): 削除時の確認プロンプトをスキップするかどうか（デフォルト: `false`）
- `close.force` (boolean): 閉じるときの確認プロンプトをスキップするかどうか（デフォルト: `false`）
- `editor` (string): 使用するエディターコマンド（例: `code`, `vim`, `emacs`）

**注意**: フラグの優先順位は以下の通りです：`--no-*` フラグ > 通常フラグ > 設定ファイル

#### --no-* フラグについて

設定ファイルで有効化したオプションをコマンド実行時に無効化できます：

- `--no-open`: `add.open=true` でも開かない
- `--no-sync`: `add.sync=true` でも同期しない
- `--no-sync-ignored`: `add.sync_ignored=true` でも gitignored ファイルを同期しない
- `--no-yes` / `--no-force`: `close.force=true` または `rm.force=true` でも確認プロンプトを表示
- `--no-branch`: `rm.branch=true` でもブランチを削除しない

```bash
# 例: config で add.open=true でも開かない
gw add --no-open feature/hoge

# 例: config で rm.branch=true でもブランチを残す
gw rm --no-branch feature/hoge
```

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

# 設定ファイルで add.open=true でも開かない（--no-open フラグ）
gw add --no-open feature/hoge

# オプションの組み合わせも可能
gw add -b --open --editor code feature/new
gw add --pr 123 --open -e vim
```

### ワークツリー一覧

```bash
gw ls
# 出力形式: <ディレクトリ名>\t<ブランチ名>\t<コミットハッシュ>\t<メインマーカー>
# ex-repo	main	a1b2c3d	(main)
# ex-repo-feature-hoge	feature/hoge	b4e5f6c
# ex-repo-fix-foo	fix/foo	c7d8e9f

# フルパスのみ出力
gw ls -p
# /path/to/ex-repo
# /path/to/ex-repo-feature-hoge
# /path/to/ex-repo-fix-foo
```

### ワークツリーの削除

```bash
# 以下はすべて同じワークツリーを指定
gw rm feature/hoge
gw rm feature-hoge
gw rm ex-repo-feature-hoge

# 複数のワークツリーを一度に削除
gw rm feature/hoge feature/fuga fix/foo

# ブランチも一緒に削除（-b/--branch オプション）
gw rm -b feature/hoge
gw rm --branch feature-hoge

# 強制削除（マージされていないブランチも削除）
gw rm -f -b feature/hoge

# 引数なしで fzf でインタラクティブに選択（Tab で複数選択可能）
gw rm
```

**注意**: ブランチ削除には以下の安全性チェックが適用されます：
- `main` または `master` ブランチは削除できません
- カレントブランチは削除できません
- マージされていないブランチは `-f`/`--force` フラグなしでは削除できません

### ワークツリーでコマンド実行

```bash
gw exec feature/hoge git status
gw exec feature-hoge npm install

# ワークツリー名を省略すると fzf で選択
gw exec git status
```

### ワークツリーへ移動

```bash
# 指定したワークツリーに移動
gw sw feature/hoge

# fzf でインタラクティブに選択
gw sw
```

### 現在のワークツリーを閉じる

```bash
# 現在のワークツリーを閉じてメインワークツリーに戻る
gw close

# 確認プロンプトをスキップして閉じる
gw close -y
gw close --yes

# ブランチも一緒に削除
gw close -b
gw close --branch

# 強制的に閉じる（マージされていないブランチも削除）
gw close -f -b
```

**注意**: `gw close` コマンドは：
- メインワークツリー（`main` または `master`）からは実行できません
- シェル統合が必要です（`gw init` のセットアップが必要）
- 設定ファイルで `close.force=true` を設定すると確認プロンプトをスキップできます

## コマンド一覧

| コマンド | エイリアス | 説明 |
|---------|-----------|------|
| `gw add <branch>` | `gw a` | ワークツリー作成 |
| `gw add` | `gw a` | 引数なしで fzf によるブランチ選択 |
| `gw add -b <branch>` | `gw a -b` | 新規ブランチ + ワークツリー作成 |
| `gw add --pr <url\|number>` | `gw a --pr`, `gw a -p` | PR ブランチのワークツリー作成 |
| `gw add --open` | `gw a --open` | ワークツリー作成後にエディターで開く |
| `gw add --no-open` | `gw a --no-open` | 設定を無視してエディターで開かない |
| `gw add --editor <cmd>` | `gw a -e` | 使用するエディターコマンドを指定 |
| `gw add --sync` | `gw a --sync` | メインワークツリーからファイルを同期 |
| `gw add --no-sync` | `gw a --no-sync` | 設定を無視してファイルを同期しない |
| `gw add --sync-ignored` | `gw a --sync-ignored` | gitignored ファイルも同期 |
| `gw add --no-sync-ignored` | `gw a --no-sync-ignored` | 設定を無視して gitignored ファイルを同期しない |
| `gw ls` | `gw l` | ワークツリー一覧表示 |
| `gw ls -p` | `gw l -p` | ワークツリーのフルパスのみ表示 |
| `gw rm [name...]` | `gw r` | ワークツリー削除（引数なしまたは複数指定可能） |
| `gw rm` | `gw r` | 引数なしで fzf による選択（Tab で複数選択可能） |
| `gw rm -b <name>` | `gw r -b` | ワークツリーとブランチを削除 |
| `gw rm --no-branch <name>` | `gw r --no-branch` | 設定を無視してブランチを削除しない |
| `gw rm --yes/-y` | `gw r -y` | 確認プロンプトをスキップ |
| `gw rm --no-yes/--no-force` | `gw r --no-yes` | 設定を無視して確認プロンプトを表示 |
| `gw exec [name] <cmd...>` | `gw e` | 対象ワークツリーでコマンド実行（引数なしで fzf） |
| `gw sw [name]` | `gw s` | 対象ワークツリーに移動（引数なしで fzf） |
| `gw close [flags]` | `gw c` | 現在のワークツリーを閉じてメインに戻る |
| `gw close -b` | `gw c -b` | ワークツリーとブランチを削除して閉じる |
| `gw close -y/--yes` | `gw c -y` | 確認プロンプトをスキップして閉じる |
| `gw close --no-yes/--no-force` | `gw c --no-yes` | 設定を無視して確認プロンプトを表示 |
| `gw fd` | `gw f` | fzf でワークツリー検索（ブランチ名を出力） |
| `gw fd -p` | `gw f -p` | fzf でワークツリー検索（フルパスを出力） |
| `gw init <shell>` | `gw i` | シェル初期化スクリプト出力 |

## 必要なツール

- `git`
- `fzf` (インタラクティブ選択用)
- `gh` (PR 連携用)

## ライセンス

MIT
