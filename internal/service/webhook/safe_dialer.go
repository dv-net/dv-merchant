package webhook

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"syscall"
	"time"
)

var ErrForbiddenDestination = errors.New("webhook destination address is not allowed")

// Ranges not covered by netip.Addr helpers (IsPrivate, IsLoopback, ...).
var forbiddenPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),     // "this" network
	netip.MustParsePrefix("100.64.0.0/10"), // carrier-grade NAT
	netip.MustParsePrefix("192.0.0.0/24"),  // IETF protocol assignments
	netip.MustParsePrefix("198.18.0.0/15"), // benchmarking
	netip.MustParsePrefix("240.0.0.0/4"),   // reserved, incl. broadcast
	netip.MustParsePrefix("64:ff9b::/96"),  // NAT64 (may map to internal IPv4)
	netip.MustParsePrefix("64:ff9b:1::/48"),
}

func isForbiddenIP(addr netip.Addr) bool {
	addr = addr.Unmap()
	if !addr.IsValid() ||
		addr.IsLoopback() ||
		addr.IsPrivate() ||
		addr.IsUnspecified() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() ||
		addr.IsInterfaceLocalMulticast() ||
		addr.IsMulticast() {
		return true
	}
	for _, p := range forbiddenPrefixes {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

func ssrfControl(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrForbiddenDestination, address)
	}
	addr, err := netip.ParseAddr(host)
	if err != nil || isForbiddenIP(addr) {
		return fmt.Errorf("%w: %s", ErrForbiddenDestination, host)
	}
	return nil
}

func newWebhookDialer(allowPrivateNetworks bool) *net.Dialer {
	d := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	if !allowPrivateNetworks {
		d.Control = ssrfControl
	}
	return d
}
