package open

import (
	"fmt"
	"os/exec"
)

// Browser はmacOSの`open`コマンドでlocalhostの指定ポートをブラウザで開く。
func Browser(port int) error {
	url := fmt.Sprintf("http://localhost:%d", port)
	return exec.Command("open", url).Run()
}
