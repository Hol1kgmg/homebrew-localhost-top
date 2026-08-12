package process

// Process はlsofで検出したlocalhost上でLISTENしているTCPプロセスを表す。
type Process struct {
	Command string
	PID     int
	User    string
	Address string // 127.0.0.1 / localhost
	Port    int
	Proto   string // TCP固定
}
