package open

import (
	"fmt"
	"os/exec"
	"strings"
)

// Browser はmacOSの`open`コマンドでlocalhostの指定ポートをブラウザで開く。
func Browser(port int) error {
	url := fmt.Sprintf("http://localhost:%d", port)
	return exec.Command("open", url).Run()
}

// Copy はmacOSの`pbcopy`コマンドで文字列をクリップボードにコピーする。
func Copy(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
