package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// lazygit風カラーパレット。
const (
	colorGreen  = lipgloss.Color("2")
	colorYellow = lipgloss.Color("3")
	colorRed    = lipgloss.Color("1")
	colorCyan   = lipgloss.Color("6")
	colorGray   = lipgloss.Color("240")
	colorWhite  = lipgloss.Color("255")
)

var (
	titleBarStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGreen)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGreen)

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorYellow)

	descStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	statusOKStyle = lipgloss.NewStyle().Foreground(colorGreen)
	errorStyle    = lipgloss.NewStyle().Foreground(colorRed).Bold(true)

	inputPromptStyle = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)

	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorCyan).
			Padding(1, 2)

	confirmPopupStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(colorRed).
				Padding(1, 3)

	confirmTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWhite).
				Background(colorRed).
				Padding(0, 1)

	confirmLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorGray)
)

func tableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorGreen).
		BorderBottom(true).
		Bold(true).
		Foreground(colorGreen)
	s.Selected = s.Selected.
		Foreground(colorWhite).
		Background(colorGreen).
		Bold(true)
	s.Cell = s.Cell.Foreground(colorWhite)
	return s
}

func hint(key, desc string) string {
	return keyStyle.Render(key) + " " + descStyle.Render(desc)
}

func (m Model) View() string {
	switch m.mode {
	case modeDetail:
		box := popupStyle.Render(
			panelTitleStyle.Render("プロセス詳細") + "\n\n" + m.detailContent,
		)
		return box + "\n" + descStyle.Render("esc/q で閉じる")
	case modeConfirmKill:
		return m.renderConfirmPopup()
	case modeHelp:
		return m.renderHelpPopup()
	}

	title := titleBarStyle.Render("● localhost-top") + "  " +
		subtitleStyle.Render(fmt.Sprintf("%d processes", len(m.visible)))

	borderColor := colorGreen
	panelTitle := "Processes"
	switch m.mode {
	case modeSearch:
		borderColor = colorCyan
		panelTitle = "Processes (search)"
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)
	if m.width > 0 {
		// Width()はborder分を含まないため、境界線(左右1文字ずつ)を除いた値を指定する。
		panelStyle = panelStyle.Width(m.width - 2)
	}
	panel := panelStyle.Render(withPanelTitle(panelTitle, borderColor, m.table.View()))

	bottom := m.renderBottom()

	return strings.Join([]string{title, panel, bottom}, "\n")
}

// withPanelTitle はボーダー枠の1行目にパネルタイトルを埋め込む（lazygit風）。
func withPanelTitle(title string, color lipgloss.Color, body string) string {
	label := lipgloss.NewStyle().Bold(true).Foreground(color).Render(" " + title + " ")
	return label + "\n" + body
}

func (m Model) renderBottom() string {
	switch m.mode {
	case modeSearch:
		return inputPromptStyle.Render("/") + m.searchInput
	case modeCommand:
		return inputPromptStyle.Render(":") + m.cmdInput
	}

	hints := strings.Join([]string{
		hint("j/k", "move"),
		hint("/", "search"),
		hint("K", "kill"),
		hint("X", "force-kill"),
		hint("enter/l", "detail"),
		hint("o", "open"),
		hint("L", "LAN link"),
		hint("s", "sort:"+m.sort.String()),
		hint("r", "reload"),
		hint(":", "cmd"),
		hint("q", "quit"),
		hint("?", "help"),
	}, "  ")

	lines := []string{hints}
	if m.status != "" {
		lines = append([]string{statusOKStyle.Render(m.status)}, lines...)
	}
	if m.err != nil {
		lines = append([]string{errorStyle.Render(fmt.Sprintf("エラー: %v", m.err))}, lines...)
	}
	return strings.Join(lines, "\n")
}

// renderConfirmPopup はkill確認を画面中央のポップアップウィンドウとして描画する。
func (m Model) renderConfirmPopup() string {
	sig := "SIGTERM"
	if m.pendingForce {
		sig = "SIGKILL"
	}

	var lines []string
	if m.pendingAll {
		lines = []string{
			confirmTitleStyle.Render(" ⚠ 全プロセスKill確認 "),
			"",
			confirmLabelStyle.Render("対象  ") + "  " + fmt.Sprintf("%d件", len(m.pendingPIDs)),
			confirmLabelStyle.Render("SIGNAL") + "  " + errorStyle.Render(sig),
			"",
			descStyle.Render("y") + " で実行 / " + descStyle.Render("n, esc") + " でキャンセル",
		}
	} else {
		lines = []string{
			confirmTitleStyle.Render(" ⚠ Kill確認 "),
			"",
			confirmLabelStyle.Render("COMMAND") + "  " + m.pendingCommand,
			confirmLabelStyle.Render("PID    ") + "  " + strconv.Itoa(m.pendingPID),
			confirmLabelStyle.Render("PORT   ") + "  " + strconv.Itoa(m.pendingPort),
			confirmLabelStyle.Render("SIGNAL ") + "  " + errorStyle.Render(sig),
			"",
			descStyle.Render("y") + " で実行 / " + descStyle.Render("n, esc") + " でキャンセル",
		}
	}

	popup := confirmPopupStyle.Render(strings.Join(lines, "\n"))

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, popup)
	}
	return popup
}

// renderHelpPopup はキーバインド・コマンド一覧を画面中央のポップアップとして描画する。
func (m Model) renderHelpPopup() string {
	row := func(kb key.Binding, desc string) string {
		return hint(strings.Join(kb.Keys(), "/"), desc)
	}

	lines := []string{
		panelTitleStyle.Render("ヘルプ"),
		"",
		row(keys.Down, "下移動") + "    " + row(keys.Up, "上移動"),
		row(keys.Top, "先頭へジャンプ") + "  " + row(keys.Bottom, "末尾へジャンプ"),
		row(keys.Search, "検索モード") + "  " + row(keys.Next, "次の検索結果") + "  " + row(keys.Prev, "前の検索結果"),
		row(keys.Kill, "kill (SIGTERM)") + "  " + row(keys.Force, "強制kill (SIGKILL)"),
		row(keys.Detail, "詳細表示") + "  " + row(keys.Open, "ブラウザで開く"),
		row(keys.LANLink, "LANアクセス用リンクを取得（0.0.0.0 bindのみ）"),
		row(keys.Sort, "ソート切替") + "  " + row(keys.Reload, "再読み込み"),
		row(keys.Command, "コマンドモード") + "  " + row(keys.Quit, "終了"),
		"",
		descStyle.Render("コマンド（:入力後）:"),
		hint(":killall", "表示中の全プロセスをkill (SIGTERM)"),
		hint(":killall!", "表示中の全プロセスを強制kill (SIGKILL)"),
		hint(":q", "終了"),
		"",
		descStyle.Render("esc / q / ? で閉じる"),
	}

	box := popupStyle.Render(strings.Join(lines, "\n"))
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}
	return box
}
