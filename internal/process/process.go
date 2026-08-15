package process

import "github.com/Hol1kgmg/homebrew-localhost-top/internal/i18n"

// Scope はプロセスがどの範囲のインターフェースでLISTENしているかを表す。
type Scope int

const (
	// ScopeLoopback は127.0.0.1/localhost/::1限定でLISTENしている状態。
	// OSレベルでloopback以外からの接続を受け付けないため、LANからはアクセス不可。
	ScopeLoopback Scope = iota
	// ScopeAll は0.0.0.0/*で全インターフェースLISTENしている状態。LANからアクセス可能。
	ScopeAll
)

func (s Scope) String() string {
	switch s {
	case ScopeAll:
		return i18n.T("scope_lan_public")
	default:
		return i18n.T("scope_local_only")
	}
}

// Process はlsofで検出したlocalhost上でLISTENしているTCPプロセスを表す。
type Process struct {
	Command string
	PID     int
	User    string
	Address string // 127.0.0.1 / localhost / 0.0.0.0 / *
	Port    int
	Proto   string // TCP固定
	Scope   Scope
}
