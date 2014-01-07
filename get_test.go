package vpngate

import (
	"net/http"
	"os"
	"testing"

	"github.com/StalkR/httpcache"
)

// client is used by tests to perform cached requests.
// If cache directory exists it is used as a persistent cache.
// Otherwise a volatile memory cache is used.
var client *http.Client

func init() {
	if _, err := os.Stat("cache"); err == nil {
		client = httpcache.NewPersistentClient("cache")
	} else {
		client = httpcache.NewVolatileClient()
	}
}

func TestGet(t *testing.T) {
	vpns, err := Get(client)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if len(vpns) == 0 {
		t.Fatalf("no VPN found")
	}
	t.Logf("Found %d VPNs", len(vpns))
}
