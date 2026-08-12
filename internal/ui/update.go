package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Hol1kgmg/homebrew-localhost-top/internal/kill"
	"github.com/Hol1kgmg/homebrew-localhost-top/internal/open"
	"github.com/Hol1kgmg/homebrew-localhost-top/internal/process"
)

type detailMsg struct {
	content string
	err     error
}

func killCmd(pid int, force bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if force {
			err = kill.Force(pid)
		} else {
			err = kill.Term(pid)
		}
		return killResultMsg{pid: pid, err: err}
	}
}

func detailCmd(pid int) tea.Cmd {
	return func() tea.Msg {
		content, err := process.Detail(pid)
		return detailMsg{content: content, err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		// タイトル行(1) + パネル境界線(2) + パネルタイトル行(1) + フッター(1) 分を差し引く。
		m.table.SetHeight(msg.Height - 5)
		m.resizeColumns(msg.Width - 4) // パネルのボーダー・パディング分を差し引く
		m.table.SetWidth(msg.Width - 4)
		return m, nil

	case tickMsg:
		return m, tea.Batch(fetchProcesses, tick())

	case processListMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		m.all = msg.processes
		m.applyFilterAndSort()
		return m, nil

	case killResultMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("kill失敗 (PID %d): %v", msg.pid, msg.err)
		} else {
			m.status = fmt.Sprintf("PID %d を終了しました", msg.pid)
		}
		return m, fetchProcesses

	case detailMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("詳細取得失敗: %v", msg.err)
			m.mode = modeNormal
			return m, nil
		}
		m.detailContent = msg.content
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeSearch:
		return m.handleSearchKey(msg)
	case modeCommand:
		return m.handleCommandKey(msg)
	case modeConfirmKill:
		return m.handleConfirmKey(msg)
	case modeDetail:
		return m.handleDetailKey(msg)
	default:
		return m.handleNormalKey(msg)
	}
}

func (m Model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch {
	case key == "q":
		return m, tea.Quit

	case key == "j" || key == "down":
		m.table.MoveDown(1)
		m.lastKey = ""
		return m, nil

	case key == "k" || key == "up":
		m.table.MoveUp(1)
		m.lastKey = ""
		return m, nil

	case key == "g":
		if m.lastKey == "g" {
			m.table.GotoTop()
			m.lastKey = ""
		} else {
			m.lastKey = "g"
		}
		return m, nil

	case key == "G":
		m.table.GotoBottom()
		m.lastKey = ""
		return m, nil

	case key == "/":
		m.mode = modeSearch
		m.searchInput = ""
		m.lastKey = ""
		return m, nil

	case key == "r":
		m.status = "再読み込み中..."
		m.lastKey = ""
		return m, fetchProcesses

	case key == "s":
		m.sort = (m.sort + 1) % 3
		m.applyFilterAndSort()
		m.status = fmt.Sprintf("ソート: %s", m.sort.String())
		m.lastKey = ""
		return m, nil

	case key == "K":
		if p, ok := m.selected(); ok {
			m.pendingPID = p.PID
			m.pendingForce = false
			m.mode = modeConfirmKill
		}
		m.lastKey = ""
		return m, nil

	case key == "X":
		if p, ok := m.selected(); ok {
			m.pendingPID = p.PID
			m.pendingForce = true
			m.mode = modeConfirmKill
		}
		m.lastKey = ""
		return m, nil

	case key == "enter" || key == "l":
		if p, ok := m.selected(); ok {
			m.mode = modeDetail
			m.detailContent = "読み込み中..."
			m.lastKey = ""
			return m, detailCmd(p.PID)
		}
		m.lastKey = ""
		return m, nil

	case key == "o":
		if p, ok := m.selected(); ok {
			if err := open.Browser(p.Port); err != nil {
				m.status = fmt.Sprintf("ブラウザで開けませんでした: %v", err)
			} else {
				m.status = fmt.Sprintf("http://localhost:%d を開きました", p.Port)
			}
		}
		m.lastKey = ""
		return m, nil

	case key == ":":
		m.mode = modeCommand
		m.cmdInput = ""
		m.lastKey = ""
		return m, nil
	}

	m.lastKey = ""
	return m, nil
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.lastQuery = m.searchInput
		m.mode = modeNormal
		m.applyFilterAndSort()
		return m, nil
	case "esc":
		m.mode = modeNormal
		return m, nil
	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}
		return m, nil
	default:
		if len(msg.String()) == 1 {
			m.searchInput += msg.String()
		}
		return m, nil
	}
}

func (m Model) handleCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		cmd := m.cmdInput
		m.mode = modeNormal
		if cmd == "q" {
			return m, tea.Quit
		}
		m.status = fmt.Sprintf("不明なコマンド: %s", cmd)
		return m, nil
	case "esc":
		m.mode = modeNormal
		return m, nil
	case "backspace":
		if len(m.cmdInput) > 0 {
			m.cmdInput = m.cmdInput[:len(m.cmdInput)-1]
		}
		return m, nil
	default:
		if len(msg.String()) == 1 {
			m.cmdInput += msg.String()
		}
		return m, nil
	}
}

func (m Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		m.mode = modeNormal
		return m, killCmd(m.pendingPID, m.pendingForce)
	case "n", "esc":
		m.mode = modeNormal
		m.status = "killをキャンセルしました"
		return m, nil
	}
	return m, nil
}

func (m Model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.mode = modeNormal
		return m, nil
	}
	return m, nil
}
