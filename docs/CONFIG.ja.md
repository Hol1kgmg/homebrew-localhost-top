# 設定ファイル

[Read in English](./CONFIG.md)

localhost-topは`~/.config/localhost-top/config.json`（`$XDG_CONFIG_HOME`が設定されている場合は`$XDG_CONFIG_HOME/localhost-top/config.json`）を読み込み、動作を設定できます。

設定ファイルが存在しない場合、または読み込み・パースに失敗した場合はすべてデフォルト値で動作します。

## 設定項目

| キー | 型 | 値 | デフォルト | 説明 |
|---|---|---|---|---|
| `language` | string | `"en"` / `"ja"` | `"en"` | UIの表示言語。`"ja"`以外（未指定・不正な値含む）はすべて英語として扱われる |

## 設定例

```json
{
  "language": "ja"
}
```

## 反映方法

設定はプロセス起動時に一度だけ読み込まれます（`main.go`起動直後）。変更した場合はlocalhost-topを再起動してください。

## 実装

- 判定ロジック: `internal/i18n/i18n.go`の`Detect()`
- パス解決: `internal/i18n/i18n.go`の`configPath()`（`XDG_CONFIG_HOME`環境変数 → 未設定時は`os.UserHomeDir()/.config`）
