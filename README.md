# localhost-top

vimキーバインドで操作する、localhost上のLISTENプロセス管理TUIツールです。`lsof`の結果をリアルタイムに一覧表示し、その場でプロセスをkillしたり、詳細を確認したり、ブラウザで開いたりできます。

![localhost-top screenshot](./src/screenshot.png)

## 特徴

- `127.0.0.1` / `localhost`（loopback限定bind）、`0.0.0.0`（全インターフェース待受）でLISTENしているTCPプロセスを対象に表示、bind範囲は`SCOPE`列で確認可能
- 全インターフェース待受（`0.0.0.0`）のプロセスは、同一LAN内の他デバイスからアクセス可能なリンクを取得してクリップボードにコピー可能
- 2秒間隔で自動リロード
- vimライクなキー操作
- kill実行時は誤操作防止のための確認プロンプト付き

## インストール

```bash
brew tap Hol1kgmg/localhost-top
brew install --cask localhost-top
```

### ソースからビルド

```bash
git clone https://github.com/Hol1kgmg/homebrew-localhost-top.git
cd homebrew-localhost-top
go build -o localhost-top .
```

## 使い方

```bash
localhost-top
```

### キーバインド

| キー | 動作 |
|---|---|
| `j` / `k` | 下移動 / 上移動 |
| `gg` / `G` | 先頭 / 末尾へジャンプ |
| `/` | 検索モード（コマンド名・ポート番号でフィルタ） |
| `n` / `N` | 検索結果の次 / 前 |
| `K` | 選択中プロセスをkill（SIGTERM、確認プロンプトあり） |
| `X` | 強制kill（SIGKILL、確認プロンプトあり） |
| `Enter` / `l` | 選択中プロセスの詳細表示（`lsof`フル出力） |
| `o` | 選択中ポートをブラウザで開く（`http://localhost:PORT`） |
| `L` | LANアクセス用リンクを取得しクリップボードにコピー（`0.0.0.0`bindのプロセスのみ。`127.0.0.1`限定bindの場合は警告のみ表示） |
| `s` | ソート切り替え（PORT → PID → USER） |
| `r` | リスト再読み込み |
| `:` | コマンドモード（`:q`で終了） |
| `q` | 終了 |
| `?` | ヘルプ表示 |

### コマンド（`:`入力後）

| コマンド | 動作 |
|---|---|
| `:q` | 終了 |
| `:killall` | 表示中の全プロセスをkill（SIGTERM、確認プロンプトあり） |
| `:killall!` | 表示中の全プロセスを強制kill（SIGKILL、確認プロンプトあり） |

`:killall` / `:killall!` は検索フィルタが有効な場合、フィルタ後に表示されているプロセスのみを対象とします。

## 前提条件

- macOS（`lsof` / `open` / `pbcopy` コマンドに依存）
- Go 1.21以上（ソースからビルドする場合）

## 開発環境のセットアップ

[mise](https://mise.jdx.dev/) がインストールされ、シェルに統合されていること。

macOS (Homebrew):

```bash
brew install mise
```

シェル統合 (zsh):

```bash
echo 'eval "$(mise activate zsh)"' >> ~/.zshrc
source ~/.zshrc
```

```bash
mise trust && mise run setup
```

gitleaks / lefthook のインストールと Git フックの設定、Go / goreleaser のインストールが一括で行われます。

### 動作確認

```bash
go run .
```

### リリース

`v*.*.*`形式のタグをpushすると、GitHub Actions（`.github/workflows/release.yml`）が自動でgoreleaserを実行します。

```bash
git tag vX.X.X
git push origin vX.X.X
```

同一リポジトリの`Casks/`ディレクトリにHomebrew Caskが自動生成・コミットされます。

## ライセンス

[MIT](./LICENSE)
