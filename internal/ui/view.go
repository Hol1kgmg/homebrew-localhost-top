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
	case modeLANGuide:
		return m.renderLANGuidePopup()
	}

	title := titleBarStyle.Render("● localhost-top") + "  " +
		subtitleStyle.Render(fmt.Sprintf("%d processes", len(m.visible)))
	if m.updateAvailable {
		title += "  " + keyStyle.Render(fmt.Sprintf("🆕 %s 利用可能 (:update)", m.latestVersion))
	}

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

	bottom := m.renderBottom()

	if m.height > 0 {
		// タイトル行(1) + パネル境界線(2) + パネルタイトル行(1) + フッター実行数分を差し引く。
		// フッターはステータス/エラーの有無で行数が変わるため、都度実測して反映する。
		footerLines := lipgloss.Height(bottom)
		m.table.SetHeight(m.height - 4 - footerLines)
	}

	panel := panelStyle.Render(withPanelTitle(panelTitle, borderColor, m.tableOrEmptyView()))

	return strings.Join([]string{title, panel, bottom}, "\n")
}

// tableOrEmptyView は該当プロセスが0件の場合に空状態メッセージを表示する。
func (m Model) tableOrEmptyView() string {
	if len(m.visible) > 0 {
		return m.table.View()
	}
	msg := "該当するプロセスがありません"
	if m.lastQuery != "" {
		msg = fmt.Sprintf("%q に一致するプロセスがありません", m.lastQuery)
	}
	return descStyle.Render(msg)
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

// lanGuideContent は127.0.0.1限定bindのプロセスに対し、0.0.0.0で
// bindし直すための代表的な起動オプション例を案内するポップアップ本文を返す。
func lanGuideContent(pid int, command string) string {
	lines := []string{
		fmt.Sprintf("PID %d (%s) はローカル限定bindのためLANから到達できません。", pid, command),
		"0.0.0.0で待受するオプションを付けて起動し直してください。",
		"",
		"代表的な起動例:",
		"  vite / react     --host 0.0.0.0",
		"  next dev         -H 0.0.0.0",
		"  rails server     -b 0.0.0.0",
		"  python -m http.server  --bind 0.0.0.0",
		"  node http.createServer().listen(port, '0.0.0.0')",
	}
	return strings.Join(lines, "\n")
}

// renderLANGuidePopup はLANアクセス不可時の起動方法案内を画面中央のポップアップとして描画する。
func (m Model) renderLANGuidePopup() string {
	box := popupStyle.Render(strings.Join([]string{
		panelTitleStyle.Render("⚠ LANアクセス不可"),
		"",
		m.lanGuideContent,
		"",
		descStyle.Render("esc/q で閉じる"),
	}, "\n"))

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}
	return box
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
		hint("esc", "検索フィルターを解除"),
		row(keys.Kill, "kill (SIGTERM)") + "  " + row(keys.Force, "強制kill (SIGKILL)"),
		row(keys.Detail, "詳細表示") + "  " + row(keys.Open, "ブラウザで開く"),
		row(keys.LANLink, "LANアクセス用リンクを取得（ローカル限定bindの場合は起動方法を案内）"),
		row(keys.Sort, "ソート切替") + "  " + row(keys.Reload, "再読み込み"),
		row(keys.Command, "コマンドモード") + "  " + row(keys.Quit, "終了"),
		"",
		descStyle.Render("コマンド（:入力後）:"),
		hint(":killall", "表示中の全プロセスをkill (SIGTERM)"),
		hint(":killall!", "表示中の全プロセスを強制kill (SIGKILL)"),
		hint(":update", "新バージョンの有無を確認"),
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
