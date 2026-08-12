Go + Bubbletea構成での実装プランです。承認をいただいてから実装に入ります。

## 実装プラン

### 1. プロジェクト構成

ツール名は `localhost-top`。このリポジトリ（`homebrew-localhost-top`）自体をソース兼tapとして使う。

```
.                              # リポジトリルート = ソース兼tap
├── main.go
├── internal/
│   ├── process/
│   │   ├── lsof.go          # lsof実行・パース
│   │   └── process.go       # Process構造体定義
│   ├── ui/
│   │   ├── model.go         # Bubbletea Model (State管理)
│   │   ├── update.go        # キー入力ハンドリング
│   │   ├── view.go          # 描画ロジック
│   │   └── keymap.go        # vimキーバインド定義
│   ├── kill/
│   │   └── kill.go          # SIGTERM/SIGKILL実行
│   └── open/
│       └── open.go          # ブラウザオープン（`open`コマンド実行）
├── Casks/
│   └── localhost-top.rb     # goreleaserが自動生成・更新（Homebrew Cask形式）
├── go.mod
├── go.sum
├── .goreleaser.yml
└── README.md
```

### 2. データ取得層（lsof連携）
- `lsof -iTCP -sTCP:LISTEN -n -P` を`os/exec`で実行
- 対象は**localhost限定**（`127.0.0.1` / `localhost`でLISTENしているプロセスのみ）。`0.0.0.0`など全インターフェース待受は対象外
- 出力をパースして以下の構造体に変換
```go
type Process struct {
    Command string
    PID     int
    User    string
    Port    int
    Proto   string // TCP/UDP
}
```
- 定期的に再取得（例: 2秒間隔でtea.Tick）してリスト更新

### 3. vimキーバインド設計

| キー | 動作 |
|---|---|
| `j` / `k` | 下移動 / 上移動 |
| `gg` / `G` | 先頭 / 末尾へジャンプ |
| `/` | 検索モード（コマンド名・ポート番号でフィルタ） |
| `n` / `N` | 検索結果の次 / 前 |
| `K` | 選択中プロセスをkill（SIGTERM） |
| `X` | 強制kill（SIGKILL） |
| `r` | リスト再読み込み |
| `Enter` / `l` | 選択中プロセスの詳細表示（ポップアップ） |
| `o` | 選択中ポートをブラウザで開く（`open http://localhost:PORT`） |
| `s` | ソート切り替え（PORT / PID / USER 順） |
| `:` | コマンドモード（`:q`で終了など） |
| `q` | 終了 |
| `?` | ヘルプ表示 |

- kill実行時は誤操作防止のため確認プロンプト（`y/n`）を必須で挟む（`confirm-kill`モードに遷移し、`y`で実行・`n`/`Esc`でキャンセル）

### 4. UI設計（Bubbletea Model）
- 状態: `normal` / `search` / `command` / `confirm-kill` / `detail` のモード管理（vimのモーダル操作を模倣）
- リスト表示は`bubbles/table`コンポーネントを使用（COMMAND, PID, USER, PORT列）
- ステータスバー下部に現在のモード・キー入力中の状態・現在のソート順を表示（vim風）
- ソートは`s`キーでPORT → PID → USERの順にトグル切り替え、ヘッダーに現在のソートキーを表示
- 詳細表示は`Enter`/`l`でポップアップ（オーバーレイ）を開き、`lsof`フル出力（CWD、FD、コマンドライン全体など）を表示。`Esc`/`q`で閉じる
- ブラウザオープン（`o`）はTCPかつHTTPと推測されるポートのみ有効化し、対象外の場合はステータスバーにエラー表示

### 5. kill処理
- Go標準の`syscall.Kill(pid, syscall.SIGTERM)` / `SIGKILL`
- 権限不足（他ユーザーのプロセス）の場合はエラーメッセージ表示（sudo再実行は範囲外とするか要検討）

### 6. 配布（brew tap）
- `goreleaser`でクロスビルド（darwin/amd64, darwin/arm64）
- このリポジトリ自体がtapを兼ねるため、`goreleaser`の`homebrew_casks`設定（`brews`は非推奨のため移行）で同一リポジトリの`Casks/`ディレクトリを直接更新
- リリースフロー: `git tag vX.X.X` → `goreleaser release` → 同一リポへのCask自動コミット
- インストール: `brew install --cask Hol1kgmg/localhost-top/localhost-top`
- `goreleaser check`でバリデーション済み、`goreleaser release --snapshot --clean --skip=publish`でクロスビルド・Cask生成のドライランを確認済み

### 7. 開発ステップ順序
1. `go.mod`初期化 + lsofパース処理単体で動作確認
2. Bubbletea基本Model（リスト表示のみ、キー操作なし）
3. vimキーバインド実装（移動・検索）
4. kill機能実装（確認プロンプト付き）
5. 詳細表示・ブラウザオープン・ソート機能実装
6. goreleaser設定（同一リポのFormula/更新）・ローカルでのbrewインストールテスト
7. `git tag`でリリース・公開

---

対象プロトコルは**TCP LISTENのみ**で確定（UDPは対象外、将来的な拡張候補として保留）。

すべての確認事項が確定しました。実装に着手可能です。

