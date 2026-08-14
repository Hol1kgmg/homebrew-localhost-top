package process

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// loopbackAddresses はloopback限定bindとみなすアドレス表記の集合。
var loopbackAddresses = map[string]bool{
	"127.0.0.1": true,
	"localhost": true,
	"[::1]":     true,
	"::1":       true,
}

// allInterfaceAddresses は全インターフェースbind（LAN到達可能）とみなすアドレス表記の集合。
var allInterfaceAddresses = map[string]bool{
	"*":       true,
	"0.0.0.0": true,
	"[::]":    true,
	"::":      true,
}

// List は現在localhost上でLISTENしているTCPプロセスの一覧を取得する。
func List() ([]Process, error) {
	cmd := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-n", "-P")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// 該当プロセスが1件もない場合、lsofはexit code 1を返す。
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 && stdout.Len() > 0 {
			// ヘッダのみの出力として継続処理する。
		} else if stdout.Len() == 0 {
			return nil, fmt.Errorf("lsof実行に失敗: %w: %s", err, stderr.String())
		}
	}

	return parse(stdout.Bytes())
}

// Detail は指定PIDの`lsof`フル出力を返す（詳細表示ポップアップ用）。
func Detail(pid int) (string, error) {
	cmd := exec.Command("lsof", "-p", strconv.Itoa(pid))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stdout.Len() == 0 {
			return "", fmt.Errorf("lsof実行に失敗: %w: %s", err, stderr.String())
		}
	}
	return stdout.String(), nil
}

func parse(output []byte) ([]Process, error) {
	var result []Process
	scanner := bufio.NewScanner(bytes.NewReader(output))
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			// ヘッダ行をスキップ。
			first = false
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		command := fields[0]
		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		user := fields[2]
		addrPort := fields[len(fields)-2]

		idx := strings.LastIndex(addrPort, ":")
		if idx < 0 {
			continue
		}
		address := addrPort[:idx]
		portStr := addrPort[idx+1:]

		var scope Scope
		switch {
		case loopbackAddresses[address]:
			scope = ScopeLoopback
		case allInterfaceAddresses[address]:
			scope = ScopeAll
		default:
			continue
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		result = append(result, Process{
			Command: command,
			PID:     pid,
			User:    user,
			Address: address,
			Port:    port,
			Proto:   "TCP",
			Scope:   scope,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
