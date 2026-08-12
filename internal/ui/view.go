package ui

import (
	"fmt"
	"strings"

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
	confirmStyle  = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorRed).
			Padding(0, 1)

	inputPromptStyle = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)

	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorCyan).
			Padding(1, 2)
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
	}

	title := titleBarStyle.Render("● localhost-top") + "  " +
		subtitleStyle.Render(fmt.Sprintf("%d processes", len(m.visible)))

	borderColor := colorGreen
	panelTitle := "Processes"
	switch m.mode {
	case modeSearch:
		borderColor = colorCyan
		panelTitle = "Processes (search)"
	case modeConfirmKill:
		borderColor = colorRed
		panelTitle = "Processes (confirm)"
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
	case modeConfirmKill:
		sig := "SIGTERM"
		if m.pendingForce {
			sig = "SIGKILL"
		}
		return confirmStyle.Render(fmt.Sprintf(" PID %d に%sを送信しますか？ (y/n) ", m.pendingPID, sig))
	}

	hints := strings.Join([]string{
		hint("j/k", "move"),
		hint("/", "search"),
		hint("K", "kill"),
		hint("X", "force-kill"),
		hint("enter/l", "detail"),
		hint("o", "open"),
		hint("s", "sort:"+m.sort.String()),
		hint("r", "reload"),
		hint(":", "cmd"),
		hint("q", "quit"),
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
