package ui

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
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

	detailContent string

	status string
	err    error

	lastKey string

	width, height int
}

// 固定幅のPID/USER/PORT/SCOPE列を除いた残りをCOMMAND列に割り当てる。
const (
	pidColWidth   = 8
	userColWidth  = 12
	portColWidth  = 8
	scopeColWidth = 10
	minCmdWidth   = 10
)

func New() Model {
	t := table.New(
		table.WithFocused(true),
	)
	t.SetStyles(tableStyles())

	m := Model{
		table: t,
		mode:  modeNormal,
		sort:  sortByPort,
	}
	m.resizeColumns(60)
	return m
}

// resizeColumns はターミナル幅に合わせてCOMMAND列の幅を再計算する。
func (m *Model) resizeColumns(tableWidth int) {
	cmdWidth := tableWidth - pidColWidth - userColWidth - portColWidth - scopeColWidth - 10 // 列間の区切り分
	if cmdWidth < minCmdWidth {
		cmdWidth = minCmdWidth
	}
	m.table.SetColumns([]table.Column{
		{Title: "COMMAND", Width: cmdWidth},
		{Title: "PID", Width: pidColWidth},
		{Title: "USER", Width: userColWidth},
		{Title: "PORT", Width: portColWidth},
		{Title: "SCOPE", Width: scopeColWidth},
	})
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchProcesses, tick())
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
