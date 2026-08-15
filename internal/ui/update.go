package ui

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdp/qrterminal/v3"

	"github.com/Hol1kgmg/homebrew-localhost-top/internal/kill"
	"github.com/Hol1kgmg/homebrew-localhost-top/internal/network"
	"github.com/Hol1kgmg/homebrew-localhost-top/internal/open"
	"github.com/Hol1kgmg/homebrew-localhost-top/internal/process"
	"github.com/Hol1kgmg/homebrew-localhost-top/internal/update"
)

type detailMsg struct {
	content string
	err     error
}

type updateCheckMsg struct {
	info update.Info
	err  error
}

// checkUpdateCmd はGitHub Releasesの最新バージョンをチェックする。
func checkUpdateCmd(version string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		info, err := update.Check(ctx, version)
		return updateCheckMsg{info: info, err: err}
	}
}

// sendSignal はforceに応じてSIGTERM/SIGKILLをpidへ送る。
func sendSignal(pid int, force bool) error {
	if force {
		return kill.Force(pid)
	}
	return kill.Term(pid)
}

func killCmd(pid int, force bool) tea.Cmd {
	return func() tea.Msg {
		return killResultMsg{pid: pid, err: sendSignal(pid, force)}
	}
}

func killAllCmd(pids []int, force bool) tea.Cmd {
	return func() tea.Msg {
		failed := 0
		for _, pid := range pids {
			if err := sendSignal(pid, force); err != nil {
				failed++
			}
		}
		return killAllResultMsg{total: len(pids), failed: failed}
	}
}

// lanLinkStatus はLANアクセス用リンクを生成しクリップボードへコピーする。
// 呼び出し側でScopeAllであることを確認済みの前提。
func lanLinkStatus(p process.Process) string {
	ip, err := network.LocalIP()
	if err != nil {
		return fmt.Sprintf("LAN側IPアドレスの取得に失敗しました: %v", err)
	}

	link := fmt.Sprintf("http://%s:%d", ip, p.Port)
	if err := open.Copy(link); err != nil {
		return fmt.Sprintf("%s （クリップボードへのコピーに失敗: %v）", link, err)
	}
	return fmt.Sprintf("%s をクリップボードにコピーしました", link)
}

// buildQRCode はLANアクセス用URLのQRコードを生成し、URL文字列と併せて返す。
// 呼び出し側でScopeAllであることを確認済みの前提。
func buildQRCode(p process.Process) (string, error) {
	ip, err := network.LocalIP()
	if err != nil {
		return "", fmt.Errorf("LAN側IPアドレスの取得に失敗しました: %w", err)
	}

	link := fmt.Sprintf("http://%s:%d", ip, p.Port)

	var buf bytes.Buffer
	qrterminal.GenerateHalfBlock(link, qrterminal.L, &buf)
	qr := strings.TrimRight(buf.String(), "\n")

	return link + "\n\n" + qr, nil
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
		// テーブルの高さはフッターの実行数（ステータス/エラーの有無で変動）に応じてView()側で都度計算する。
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

	case statusClearMsg:
		if msg.token == m.statusToken {
			m.status = ""
		}
		return m, nil

	case killResultMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("kill失敗 (PID %d): %v", msg.pid, msg.err)
		} else {
			m.status = fmt.Sprintf("PID %d を終了しました", msg.pid)
		}
		return m, fetchProcesses

	case killAllResultMsg:
		succeeded := msg.total - msg.failed
		if msg.failed > 0 {
			m.status = fmt.Sprintf("%d件中%d件を終了しました（%d件失敗）", msg.total, succeeded, msg.failed)
		} else {
			m.status = fmt.Sprintf("%d件を終了しました", succeeded)
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

	case updateCheckMsg:
		if msg.err != nil {
			if m.status == "アップデートを確認中..." {
				m.status = fmt.Sprintf("アップデート確認に失敗しました: %v", msg.err)
			}
			return m, nil
		}
		if msg.info.Available {
			m.updateAvailable = true
			m.latestVersion = msg.info.Latest
			m.status = fmt.Sprintf("新しいバージョン %s が利用可能です", msg.info.Latest)
		} else if m.status == "アップデートを確認中..." {
			m.status = "最新バージョンです"
		}
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
	case modeHelp:
		return m.handleHelpKey(msg)
	case modeLANGuide:
		return m.handleLANGuideKey(msg)
	case modeQRCode:
		return m.handleQRCodeKey(msg)
	default:
		return m.handleNormalKey(msg)
	}
}

// startKillConfirm はK/Xキー共通のkill確認ポップアップ準備処理。
func (m *Model) startKillConfirm(p process.Process, force bool) {
	m.pendingPID = p.PID
	m.pendingForce = force
	m.pendingCommand = p.Command
	m.pendingPort = p.Port
	m.pendingAll = false
	m.mode = modeConfirmKill
}

func (m Model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	prevKey := m.lastKey
	m.lastKey = ""

	switch key {
	case "q":
		return m, tea.Quit

	case "esc":
		if m.lastQuery != "" {
			m.searchInput = ""
			m.lastQuery = ""
			m.applyFilterAndSort()
			m.status = "検索フィルターを解除しました"
		}
		return m, nil

	case "j", "down":
		m.table.MoveDown(1)
		return m, nil

	case "k", "up":
		m.table.MoveUp(1)
		return m, nil

	case "g":
		if prevKey == "g" {
			m.table.GotoTop()
		} else {
			m.lastKey = "g"
		}
		return m, nil

	case "G":
		m.table.GotoBottom()
		return m, nil

	case "/":
		m.mode = modeSearch
		m.searchInput = ""
		return m, nil

	case "r":
		clearCmd := m.setStatus("再読み込み中...")
		return m, tea.Batch(fetchProcesses, clearCmd)

	case "s":
		m.sort = (m.sort + 1) % 3
		m.refreshColumnTitles()
		m.applyFilterAndSort()
		m.status = fmt.Sprintf("ソート: %s", m.sort.String())
		return m, nil

	case "K":
		if p, ok := m.selected(); ok {
			m.startKillConfirm(p, false)
		}
		return m, nil

	case "X":
		if p, ok := m.selected(); ok {
			m.startKillConfirm(p, true)
		}
		return m, nil

	case "enter", "l":
		if p, ok := m.selected(); ok {
			m.mode = modeDetail
			m.detailContent = "読み込み中..."
			return m, detailCmd(p.PID)
		}
		return m, nil

	case "o":
		if p, ok := m.selected(); ok {
			if err := open.Browser(p.Port); err != nil {
				m.status = fmt.Sprintf("ブラウザで開けませんでした: %v", err)
			} else {
				m.status = fmt.Sprintf("http://localhost:%d を開きました", p.Port)
			}
		}
		return m, nil

	case "L":
		if p, ok := m.selected(); ok {
			if p.Scope != process.ScopeAll {
				m.lanGuideContent = lanGuideContent(p.PID, p.Command)
				m.mode = modeLANGuide
			} else {
				m.status = lanLinkStatus(p)
			}
		}
		return m, nil

	case "Q":
		if p, ok := m.selected(); ok {
			if p.Scope != process.ScopeAll {
				m.lanGuideContent = lanGuideContent(p.PID, p.Command)
				m.mode = modeLANGuide
			} else if content, err := buildQRCode(p); err != nil {
				m.status = err.Error()
			} else {
				m.qrCodeContent = content
				m.mode = modeQRCode
			}
		}
		return m, nil

	case ":":
		m.mode = modeCommand
		m.cmdInput = ""
		return m, nil

	case "?":
		m.mode = modeHelp
		return m, nil
	}

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
		m.searchInput = ""
		m.lastQuery = ""
		m.applyFilterAndSort()
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
		switch cmd {
		case "q":
			return m, tea.Quit
		case "killall", "killall!":
			if len(m.visible) == 0 {
				m.status = "killするプロセスがありません"
				return m, nil
			}
			pids := make([]int, len(m.visible))
			for i, p := range m.visible {
				pids[i] = p.PID
			}
			m.pendingAll = true
			m.pendingForce = cmd == "killall!"
			m.pendingPIDs = pids
			m.mode = modeConfirmKill
			return m, nil
		case "update":
			if m.version == "dev" {
				m.status = "開発ビルドのためアップデート確認をスキップしました"
				return m, nil
			}
			m.status = "アップデートを確認中..."
			return m, checkUpdateCmd(m.version)
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
		if m.pendingAll {
			return m, killAllCmd(m.pendingPIDs, m.pendingForce)
		}
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

func (m Model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "?":
		m.mode = modeNormal
		return m, nil
	}
	return m, nil
}

func (m Model) handleLANGuideKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.mode = modeNormal
		return m, nil
	}
	return m, nil
}

func (m Model) handleQRCodeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.mode = modeNormal
		return m, nil
	}
	return m, nil
}
