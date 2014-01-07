// Binary list shows VPNs of http://www.vpngate.net with their OpenVPN config.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/StalkR/vpngate"
)

func main() {
	vpns, err := vpngate.Get(http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}
	for _, vpn := range vpns {
		fmt.Printf("# %s (%s)\n", vpn.Hostname, vpn.Country)
		fmt.Println(vpn.OpenVPN())
		fmt.Println()
	}
}
