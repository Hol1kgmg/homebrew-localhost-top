// Package i18n はlocalhost-topのUI文言を言語ごとに切り替えるための最小限の仕組みを提供する。
package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Lang はUIの表示言語を表す。
type Lang int

const (
	EN Lang = iota
	JA
)

var current = EN

// Current は現在設定されている表示言語を返す。
func Current() Lang {
	return current
}

// SetLang は表示言語を設定する。main起動時に一度だけ呼び出すことを想定する。
func SetLang(l Lang) {
	current = l
}

// configDirName / configFileName は設定ファイルの配置場所。
// $XDG_CONFIG_HOME（未設定時は~/.config）直下のアプリ名サブディレクトリに配置する。
const (
	configDirName  = "localhost-top"
	configFileName = "config.json"
)

type config struct {
	Language string `json:"language"`
}

// configPath は設定ファイルの絶対パスを返す。取得できない場合は空文字を返す。
func configPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, configDirName, configFileName)
}

// Detect は設定ファイル（~/.config/localhost-top/config.json の"language"キー、"en"または"ja"）から
// 表示言語を決定する。ファイルが存在しない・読み取れない・値が不正な場合は英語をデフォルトとする。
func Detect() Lang {
	path := configPath()
	if path == "" {
		return EN
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return EN
	}

	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return EN
	}

	if cfg.Language == "ja" {
		return JA
	}
	return EN
}

type message struct {
	En string
	Ja string
}

var catalog = map[string]message{
	// internal/process
	"scope_lan_public": {En: "LAN public 🌐", Ja: "LAN公開 🌐"},
	"scope_local_only": {En: "local only", Ja: "local限定"},

	// internal/network
	"lan_ip_fetch_failed": {En: "Failed to get LAN-side IP address", Ja: "LAN側IPアドレスの取得に失敗しました"},

	// internal/update
	"github_api_request_failed": {En: "GitHub API request failed (status %d)", Ja: "GitHub APIへのリクエストが失敗しました (status %d)"},
	"invalid_version_format":    {En: "invalid version format: %q", Ja: "不正なバージョン形式です: %q"},

	// main
	"startup_error": {En: "Error: %v", Ja: "エラー: %v"},

	// internal/ui/update.go
	"lan_ip_fetch_failed_with_err": {En: "Failed to get LAN-side IP address: %v", Ja: "LAN側IPアドレスの取得に失敗しました: %v"},
	"clipboard_copy_failed":        {En: "%s (failed to copy to clipboard: %v)", Ja: "%s （クリップボードへのコピーに失敗: %v）"},
	"clipboard_copied":             {En: "Copied %s to clipboard", Ja: "%s をクリップボードにコピーしました"},
	"kill_failed":                  {En: "Kill failed (PID %d): %v", Ja: "kill失敗 (PID %d): %v"},
	"kill_succeeded":               {En: "Terminated PID %d", Ja: "PID %d を終了しました"},
	"kill_all_result_partial":      {En: "Terminated %d of %d (%d failed)", Ja: "%d件中%d件を終了しました（%d件失敗）"},
	"kill_all_result_success":      {En: "Terminated %d processes", Ja: "%d件を終了しました"},
	"detail_fetch_failed":          {En: "Failed to fetch details: %v", Ja: "詳細取得失敗: %v"},
	"checking_update":              {En: "Checking for updates...", Ja: "アップデートを確認中..."},
	"update_check_failed":          {En: "Failed to check for updates: %v", Ja: "アップデート確認に失敗しました: %v"},
	"new_version_available":        {En: "New version %s is available", Ja: "新しいバージョン %s が利用可能です"},
	"up_to_date":                   {En: "You're up to date", Ja: "最新バージョンです"},
	"search_filter_cleared":        {En: "Search filter cleared", Ja: "検索フィルターを解除しました"},
	"reloading":                    {En: "Reloading...", Ja: "再読み込み中..."},
	"sort_label":                   {En: "Sort: %s", Ja: "ソート: %s"},
	"loading":                      {En: "Loading...", Ja: "読み込み中..."},
	"browser_open_failed":          {En: "Failed to open browser: %v", Ja: "ブラウザで開けませんでした: %v"},
	"browser_opened":               {En: "Opened http://localhost:%d", Ja: "http://localhost:%d を開きました"},
	"no_processes_to_kill":         {En: "No processes to kill", Ja: "killするプロセスがありません"},
	"dev_build_skip_update":        {En: "Skipped update check for dev build", Ja: "開発ビルドのためアップデート確認をスキップしました"},
	"unknown_command":              {En: "Unknown command: %s", Ja: "不明なコマンド: %s"},
	"kill_cancelled":               {En: "Kill cancelled", Ja: "killをキャンセルしました"},

	// internal/ui/view.go
	"update_available_title":      {En: "🆕 %s available (:update)", Ja: "🆕 %s 利用可能 (:update)"},
	"no_matching_processes":       {En: "No matching processes", Ja: "該当するプロセスがありません"},
	"no_matching_processes_query": {En: "No processes matching %q", Ja: "%q に一致するプロセスがありません"},
	"error_prefix":                {En: "Error: %v", Ja: "エラー: %v"},
	"size_warning_title":          {En: "⚠ Display area too small", Ja: "⚠ 表示エリア不足"},
	"size_warning_desc1":          {En: "Widen your terminal until the bar below", Ja: "下の帯が折り返さず1行に収まるまで"},
	"size_warning_desc2":          {En: "fits on a single line without wrapping", Ja: "ターミナルを広げてください"},
	"size_warning_height":         {En: "%d more row(s) needed", Ja: "縦もあと%d行足りません"},
	"close_hint":                  {En: "esc/q to close", Ja: "esc/q で閉じる"},
	"detail_title":                {En: "Process Detail", Ja: "プロセス詳細"},
	"detail_indicator":            {En: "(%d/%d lines)", Ja: "(%d/%d行)"},
	"detail_scroll_hint":          {En: "j/k scroll vertical  h/l scroll horizontal  esc/q to close", Ja: "j/k 縦スクロール  h/l 横スクロール  esc/q で閉じる"},
	"confirm_all_title":           {En: " ⚠ Confirm Kill All ", Ja: " ⚠ 全プロセスKill確認 "},
	"confirm_target_label":        {En: "Target ", Ja: "対象  "},
	"confirm_target_count":        {En: "%d process(es)", Ja: "%d件"},
	"confirm_execute":             {En: "execute", Ja: "実行"},
	"confirm_cancel":              {En: "cancel", Ja: "キャンセル"},
	"confirm_title":               {En: " ⚠ Confirm Kill ", Ja: " ⚠ Kill確認 "},
	"lan_guide_unreachable":       {En: "PID %d (%s) is bound to localhost only and unreachable from the LAN.", Ja: "PID %d (%s) はローカル限定bindのためLANから到達できません。"},
	"lan_guide_rebind":            {En: "Restart it with an option to listen on 0.0.0.0.", Ja: "0.0.0.0で待受するオプションを付けて起動し直してください。"},
	"lan_guide_examples_label":    {En: "Common examples:", Ja: "代表的な起動例:"},
	"lan_guide_unreachable_title": {En: "⚠ LAN access unavailable", Ja: "⚠ LANアクセス不可"},
	"qr_title":                    {En: "LAN access QR code", Ja: "LANアクセス用QRコード"},
	"help_title":                  {En: "Help", Ja: "ヘルプ"},
	"help_move_down":              {En: "move down", Ja: "下移動"},
	"help_move_up":                {En: "move up", Ja: "上移動"},
	"help_jump_top":               {En: "jump to top", Ja: "先頭へジャンプ"},
	"help_jump_bottom":            {En: "jump to bottom", Ja: "末尾へジャンプ"},
	"help_search_mode":            {En: "search mode", Ja: "検索モード"},
	"help_clear_search":           {En: "clear search filter", Ja: "検索フィルターを解除"},
	"help_kill":                   {En: "kill (SIGTERM)", Ja: "kill (SIGTERM)"},
	"help_force_kill":             {En: "force kill (SIGKILL)", Ja: "強制kill (SIGKILL)"},
	"help_detail":                 {En: "show detail (j/k vertical, h/l horizontal scroll)", Ja: "詳細表示（j/k縦・h/l横スクロール）"},
	"help_open_browser":           {En: "open in browser", Ja: "ブラウザで開く"},
	"help_lan_link":               {En: "get LAN access link", Ja: "LANアクセス用リンクを取得"},
	"help_lan_qr":                 {En: "show LAN access link as QR code", Ja: "LANアクセス用リンクをQRコード表示"},
	"help_sort_toggle":            {En: "toggle sort", Ja: "ソート切替"},
	"help_reload":                 {En: "reload", Ja: "再読み込み"},
	"help_command_mode":           {En: "command mode", Ja: "コマンドモード"},
	"help_quit":                   {En: "quit", Ja: "終了"},
	"help_command_section":        {En: "Commands (after :):", Ja: "コマンド（:入力後）:"},
	"help_killall":                {En: "kill all visible processes (SIGTERM)", Ja: "表示中の全プロセスをkill (SIGTERM)"},
	"help_killall_force":          {En: "force kill all visible processes (SIGKILL)", Ja: "表示中の全プロセスを強制kill (SIGKILL)"},
	"help_update_check":           {En: "check for a new version", Ja: "新バージョンの有無を確認"},
	"help_close_hint":             {En: "esc / q / ? to close", Ja: "esc / q / ? で閉じる"},
}

// T は指定キーに対応する現在の表示言語のメッセージを返す。
// テンプレート文字列（%v等）の場合はfmt.Sprintf等と組み合わせて使う。
func T(key string) string {
	msg, ok := catalog[key]
	if !ok {
		return key
	}
	if current == JA {
		return msg.Ja
	}
	return msg.En
}
