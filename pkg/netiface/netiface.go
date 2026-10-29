package netiface

import (
	"context"
	"net"
	"net/netip"
	"os"
	"strings"
)

type Address struct {
	IP      string `json:"ip"`
	Private bool   `json:"private"`
} //	@name	NetworkInterfaceAddress

type Interface struct {
	Name      string    `json:"name"`
	Addresses []Address `json:"addresses"`
} //	@name	NetworkInterface

var servicePrefixes = []string{
	"docker", "br-", "veth", "virbr", "vnet", "cni", "flannel",
	"cali", "weave", "kube", "lxc", "lxdbr", "podman", "vmnet", "vboxnet",
	"dummy", "sit", "ip6tnl", "gre", "tunl", "erspan",
}

func isServiceInterface(iface net.Interface) bool {
	if iface.Flags&net.FlagUp == 0 ||
		iface.Flags&net.FlagLoopback != 0 ||
		iface.Flags&net.FlagRunning == 0 {
		return true
	}
	for _, p := range servicePrefixes {
		if strings.HasPrefix(iface.Name, p) {
			return true
		}
	}
	if _, err := os.Stat("/sys/class/net/" + iface.Name + "/bridge"); err == nil {
		return true
	}
	return false
}

var unusablePrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("192.0.0.0/24"),    // IETF protocol assignments
	netip.MustParsePrefix("192.0.2.0/24"),    // TEST-NET-1
	netip.MustParsePrefix("198.18.0.0/15"),   // benchmarks
	netip.MustParsePrefix("198.51.100.0/24"), // TEST-NET-2
	netip.MustParsePrefix("203.0.113.0/24"),  // TEST-NET-3
	netip.MustParsePrefix("240.0.0.0/4"),     // reserved addresses
	netip.MustParsePrefix("2001:db8::/32"),   // IPv6
}

var cgnat = netip.MustParsePrefix("100.64.0.0/10") // Carrier-grade NAT, Tailscale

func classify(ip netip.Addr) (usable, private bool) {
	ip = ip.Unmap() // ::ffff:1.2.3.4 -> 1.2.3.4

	if !ip.IsValid() ||
		ip.IsUnspecified() || // 0.0.0.0, ::
		ip.IsLoopback() || // 127.0.0.0/8, ::1
		ip.IsLinkLocalUnicast() || // 169.254.0.0/16, fe80::/10
		ip.IsMulticast() ||
		ip.IsInterfaceLocalMulticast() {
		return false, false
	}
	for _, p := range unusablePrefixes {
		if p.Contains(ip) {
			return false, false
		}
	}

	return true, ip.IsPrivate() || cgnat.Contains(ip)
}

func List() ([]Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	res := make([]Interface, 0, len(ifaces))
	for _, iface := range ifaces {
		if isServiceInterface(iface) {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		item := Interface{Name: iface.Name}
		for _, a := range addrs {
			prefix, err := netip.ParsePrefix(a.String()) // "192.168.1.5/24"
			if err != nil {
				continue
			}
			usable, private := classify(prefix.Addr())
			if !usable {
				continue
			}
			item.Addresses = append(item.Addresses, Address{
				IP:      prefix.Addr().String(),
				Private: private,
			})
		}
		if len(item.Addresses) > 0 {
			res = append(res, item)
		}
	}
	return res, nil
}

func Exists(name string) bool {
	list, err := List()
	if err != nil {
		return false
	}
	for _, i := range list {
		if i.Name == name {
			return true
		}
	}
	return false
}

// DefaultName returns the interface the system uses for internet traffic (default route).
// No packets are sent: dialing UDP only picks a route. Returns "" if it can't be determined.
func DefaultName(ctx context.Context) string {
	conn, err := (&net.Dialer{}).DialContext(ctx, "udp", "1.1.1.1:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	udpAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return ""
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			if ipNet, ok := a.(*net.IPNet); ok && ipNet.IP.Equal(udpAddr.IP) {
				return iface.Name
			}
		}
	}
	return ""
}
