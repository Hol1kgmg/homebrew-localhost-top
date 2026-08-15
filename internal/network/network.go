package network

import (
	"errors"
	"net"

	"github.com/Hol1kgmg/homebrew-localhost-top/internal/i18n"
)

// LocalIP は同一LAN内の他デバイスから到達可能な自機のIPv4アドレスを返す。
// 実際に通信は行わず、UDPソケットの経路解決を利用して送信元アドレスを取得する。
func LocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", errors.New(i18n.T("lan_ip_fetch_failed"))
	}
	defer conn.Close()

	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return "", errors.New(i18n.T("lan_ip_fetch_failed"))
	}
	return addr.IP.String(), nil
}
