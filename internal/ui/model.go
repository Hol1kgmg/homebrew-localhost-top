package ui

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Hol1kgmg/homebrew-localhost-top/internal/process"
)

type mode int

const (
	modeNormal mode = iota
	modeSearch
	modeCommand
	modeConfirmKill
	modeDetail
	modeHelp
	modeLANGuide
	modeQRCode
)

type sortKey int

const (
	sortByPort sortKey = iota
	sortByPID
	sortByUser
)

func (s sortKey) String() string {
	switch s {
	case sortByPort:
		return "PORT"
	case sortByPID:
		return "PID"
	case sortByUser:
		return "USER"
	default:
		return ""
	}
}

const refreshInterval = 2 * time.Second

type Model struct {
	table   table.Model
	all     []process.Process
	visible []process.Process

	mode mode
	sort sortKey

	searchInput string
	lastQuery   string

	cmdInput string

	pendingPID     int
	pendingForce   bool
	pendingCommand string
	pendingPort    int
	pendingAll     bool
	pendingPIDs    []int

	detailViewport viewport.Model

	lanGuideContent string
	qrCodeContent   string

	status      string
	statusToken int
	err         error

	lastKey string

	width, height int
	cmdColWidth   int

	version         string
	updateAvailable bool
	latestVersion   string
}

// 固定幅のPID/USER/PORT/SCOPE列を除いた残りをCOMMAND列に割り当てる。
const (
	pidColWidth   = 8
	userColWidth  = 12
	portColWidth  = 8
	scopeColWidth = 14
	minCmdWidth   = 10
)

func New(version string) Model {
	t := table.New(
		table.WithFocused(true),
	)
	t.SetStyles(tableStyles())

	m := Model{
		table:          t,
		mode:           modeNormal,
		sort:           sortByPort,
		version:        version,
		detailViewport: viewport.New(0, 0),
	}
	m.resizeColumns(60)
	return m
}

// detailPopupMargin はターミナル端から詳細ポップアップまでの余白。
const detailPopupMargin = 6

// detailPopupChromeLines はポップアップ内でviewport以外が使う行数（タイトル+空行+空行+フッター）。
const detailPopupChromeLines = 4

// resizeDetailViewport はターミナルサイズに合わせて詳細ポップアップのviewportサイズを再計算する。
func (m *Model) resizeDetailViewport() {
	w := m.width - detailPopupMargin - popupStyle.GetHorizontalFrameSize()
	if w < 20 {
		w = 20
	}
	h := m.height - detailPopupMargin - popupStyle.GetVerticalFrameSize() - detailPopupChromeLines
	if h < 3 {
		h = 3
	}
	m.detailViewport.Width = w
	m.detailViewport.Height = h
}

// setDetailContent は詳細ポップアップの本文をセットし、スクロール位置を先頭に戻す。
// lsofの列整列を保つため折り返しは行わず、viewport幅を超える行は横スクロールで参照する。
func (m *Model) setDetailContent(raw string) {
	m.resizeDetailViewport()
	m.detailViewport.SetContent(raw)
	m.detailViewport.GotoTop()
	m.detailViewport.SetXOffset(0)
}

// resizeColumns はターミナル幅に合わせてCOMMAND列の幅を再計算する。
func (m *Model) resizeColumns(tableWidth int) {
	cmdWidth := tableWidth - pidColWidth - userColWidth - portColWidth - scopeColWidth - 10 // 列間の区切り分
	if cmdWidth < minCmdWidth {
		cmdWidth = minCmdWidth
	}
	m.cmdColWidth = cmdWidth
	m.refreshColumnTitles()
}

// columnTitle は現在のソート対象列にインジケーターを付与する。
func (m *Model) columnTitle(base string, key sortKey) string {
	if m.sort == key {
		return base + " ▲"
	}
	return base
}

// refreshColumnTitles はソート状態を反映して列見出しを再構築する（幅は変えない）。
func (m *Model) refreshColumnTitles() {
	m.table.SetColumns([]table.Column{
		{Title: "COMMAND", Width: m.cmdColWidth},
		{Title: m.columnTitle("PID", sortByPID), Width: pidColWidth},
		{Title: m.columnTitle("USER", sortByUser), Width: userColWidth},
		{Title: m.columnTitle("PORT", sortByPort), Width: portColWidth},
		{Title: "SCOPE", Width: scopeColWidth},
	})
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{fetchProcesses, tick()}
	if m.version != "dev" {
		cmds = append(cmds, checkUpdateCmd(m.version))
	}
	return tea.Batch(cmds...)
}

func tick() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

type processListMsg struct {
	processes []process.Process
	err       error
}

// statusDisplayDuration はステータスメッセージの最低表示時間。
// これより早く処理が終わっても、この時間が経過するまでは表示し続ける。
const statusDisplayDuration = 1500 * time.Millisecond

type statusClearMsg struct {
	token int
}

// clearStatusAfter は指定時間後にstatusClearMsgを発行する。
// tokenが発行時点のm.statusTokenと一致する場合のみクリアすることで、
// その間に新しいステータスが設定された場合に誤って消してしまうのを防ぐ。
func clearStatusAfter(token int, d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return statusClearMsg{token: token}
	})
}

// setStatus はステータスメッセージを設定し、一定時間後に自動的にクリアするコマンドを返す。
func (m *Model) setStatus(s string) tea.Cmd {
	m.statusToken++
	m.status = s
	return clearStatusAfter(m.statusToken, statusDisplayDuration)
}

type killResultMsg struct {
	pid int
	err error
}

type killAllResultMsg struct {
	total  int
	failed int
}

func fetchProcesses() tea.Msg {
	list, err := process.List()
	return processListMsg{processes: list, err: err}
}

func (m *Model) applyFilterAndSort() {
	filtered := m.all
	if m.lastQuery != "" {
		q := strings.ToLower(m.lastQuery)
		filtered = make([]process.Process, 0, len(m.all))
		for _, p := range m.all {
			if strings.Contains(strings.ToLower(p.Command), q) ||
				strings.Contains(strconv.Itoa(p.Port), q) {
				filtered = append(filtered, p)
			}
		}
	}

	sorted := make([]process.Process, len(filtered))
	copy(sorted, filtered)
	switch m.sort {
	case sortByPort:
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Port < sorted[j].Port })
	case sortByPID:
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].PID < sorted[j].PID })
	case sortByUser:
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].User < sorted[j].User })
	}

	m.visible = sorted

	rows := make([]table.Row, len(sorted))
	for i, p := range sorted {
		rows[i] = table.Row{p.Command, strconv.Itoa(p.PID), p.User, strconv.Itoa(p.Port), p.Scope.String()}
	}
	m.table.SetRows(rows)
}

func (m Model) selected() (process.Process, bool) {
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.visible) {
		return process.Process{}, false
	}
	return m.visible[idx], true
}
