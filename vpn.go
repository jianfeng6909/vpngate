/*
Package vpngate implements parsing of http://www.vpngate.net list
of VPN servers and generation of OpenVPN config.
*/
package vpngate

import (
	"fmt"
	"time"
)

// A VPN represents a vpngate server.
type VPN struct {
	Hostname     string
	Score        int
	Ping         time.Duration
	Speed        int    // bps
	Country      string // e.g. Japan
	CountryShort string // e.g. JP
	Sessions     int    // Currently active
	Uptime       time.Duration
	Users        int // Total users
	Traffic      int // Total traffic
	LogType      string
	Operator     string
	Message      string

	// Extracted from OpenVPN config.
	Proto  string
	IP     string
	Port   int
	Cipher string
	Auth   string
	CA     string
	Cert   string
	Key    string
}

// OpenVPN config of this VPN.
func (v *VPN) OpenVPN() string {
	return fmt.Sprintf(`dev tun
proto %s
remote %s %d
cipher %s
auth %s
resolv-retry infinite
nobind
persist-key
persist-tun
client
verb 3
<ca>
%s
</ca>
<cert>
%s
</cert>
<key>
%s
</key>`, v.Proto, v.IP, v.Port, v.Cipher, v.Auth, v.CA, v.Cert, v.Key)
}
