package utils

import (
	"context"
	"net"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func newDialContext(iface string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	d := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	if iface == "" {
		return d.DialContext
	}
	d.Control = func(_, _ string, c syscall.RawConn) error {
		var serr error
		if err := c.Control(func(fd uintptr) {
			serr = unix.BindToDevice(int(fd), iface)
		}); err != nil {
			return err
		}
		return serr
	}
	return d.DialContext
}
