# Configuration file

[日本語版はこちら](./CONFIG.ja.md)

localhost-top reads `~/.config/localhost-top/config.json` (or `$XDG_CONFIG_HOME/localhost-top/config.json` if `$XDG_CONFIG_HOME` is set) to configure its behavior.

If the config file does not exist, or fails to be read or parsed, localhost-top falls back to default values for everything.

## Options

| Key | Type | Values | Default | Description |
|---|---|---|---|---|
| `language` | string | `"en"` / `"ja"` | `"en"` | UI display language. Anything other than `"ja"` (including unset or invalid values) is treated as English |

## Example

```json
{
  "language": "ja"
}
```

## When it takes effect

The config is read once at process startup (right after `main.go` starts). Restart localhost-top after changing it.

## Implementation

- Detection logic: `Detect()` in `internal/i18n/i18n.go`
- Path resolution: `configPath()` in `internal/i18n/i18n.go` (`XDG_CONFIG_HOME` env var, falling back to `os.UserHomeDir()/.config`)
