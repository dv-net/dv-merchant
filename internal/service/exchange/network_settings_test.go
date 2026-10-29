package exchange

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
)

func TestRestoreMaskedPassword(t *testing.T) {
	saved := pgtype.Text{String: "socks5://user:s3cret@203.0.113.10:1080", Valid: true}

	tests := []struct {
		name  string
		input string
		saved pgtype.Text
		want  string
	}{
		{"masked url unchanged", "socks5://user:xxxxx@203.0.113.10:1080", saved, "socks5://user:s3cret@203.0.113.10:1080"},
		{"masked password, new host", "socks5://user:xxxxx@198.51.100.7:1081", saved, "socks5://user:s3cret@198.51.100.7:1081"},
		{"new password", "socks5://user:other@203.0.113.10:1080", saved, "socks5://user:other@203.0.113.10:1080"},
		{"other user keeps mask", "socks5://admin:xxxxx@203.0.113.10:1080", saved, "socks5://admin:xxxxx@203.0.113.10:1080"},
		{"nothing saved", "socks5://user:xxxxx@203.0.113.10:1080", pgtype.Text{}, "socks5://user:xxxxx@203.0.113.10:1080"},
		{"no credentials", "http://203.0.113.10:3128", saved, "http://203.0.113.10:3128"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := utils.ParseProxyURL(tt.input)
			require.NoError(t, err)
			restoreMaskedPassword(u, tt.saved)
			require.Equal(t, tt.want, u.String())
		})
	}
}

func TestRedactProxyURL(t *testing.T) {
	require.Equal(t, "socks5://user:xxxxx@203.0.113.10:1080", redactProxyURL("socks5://user:s3cret@203.0.113.10:1080"))
	require.Equal(t, "http://203.0.113.10:3128", redactProxyURL("http://203.0.113.10:3128"))
}
