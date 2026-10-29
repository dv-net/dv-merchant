//go:build !linux

package utils

import (
	"context"
	"fmt"
	"net"
	"time"
)

func newDialContext(iface string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	base := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	if iface == "" {
		return base.DialContext
	}

	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		localIP, err := interfaceIP(iface)
		if err != nil {
			return nil, err
		}
		d := *base
		d.LocalAddr = &net.TCPAddr{IP: localIP}
		return d.DialContext(ctx, network, addr)
	}
}

func interfaceIP(name string) (net.IP, error) {
	ifc, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("network interface %q: %w", name, err)
	}
	addrs, err := ifc.Addrs()
	if err != nil {
		return nil, fmt.Errorf("network interface %q addresses: %w", name, err)
	}

	var v6 net.IP
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() || ipNet.IP.IsLinkLocalUnicast() {
			continue
		}
		if ip4 := ipNet.IP.To4(); ip4 != nil {
			return ip4, nil
		}
		if v6 == nil {
			v6 = ipNet.IP
		}
	}
	if v6 != nil {
		return v6, nil
	}
	return nil, fmt.Errorf("network interface %q has no usable address", name)
}
