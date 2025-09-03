package hash_test

import (
	"testing"

	"github.com/dv-net/dv-merchant/internal/tools/hash"

	"github.com/stretchr/testify/require"
)

func Test_ConnectionHash(t *testing.T) {
	exchangeSlug := "binance"
	input := []string{"foo", "bar", "baz"}
	expectedString := "binance_97df3588b5a3f24babc3851b372f0ba71a9dcdded43b14b9d06961bfc1707d9d"
	hashed, err := hash.SHA256ConnectionHash(exchangeSlug, input...)
	require.NoError(t, err)
	require.Equal(t, hashed, expectedString)
}
