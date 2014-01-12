package vpngate

import (
	"net/http"
)

// CSV_URL is the URL of all vpngate servers in CSV format.
const CSV_URL = "http://www.vpngate.net/api/iphone/"

// Get obtains obtains a map of all vpngate servers.
func Get(c *http.Client) (map[string]VPN, error) {
	resp, err := c.Get(CSV_URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return parseList(resp.Body)
}
