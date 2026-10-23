package kraken

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// Example from Kraken REST API docs ("Authentication" section).
func TestSigner_Sign(t *testing.T) {
	signer, err := NewSigner("kQH5HW/8p1uGOVjbgWA7FunAmGO8lsSUXNsu3eow76sz84Q18fWxnyRzBHCd3pd5nE9qa99HAZtuZuj6F1huXg==")
	require.NoError(t, err)

	sign, err := signer.Sign(
		"/0/private/AddOrder",
		"1616492376594",
		"nonce=1616492376594&ordertype=limit&pair=XBTUSD&price=37500&type=buy&volume=1.25",
	)
	require.NoError(t, err)
	require.Equal(t, "4/dpxb3iT4tp/ZCVEwSnEsLxx0bqyhLpdfOpc6fn7OR8+UClSV5n9E6aSS8MPtnRfp32bAb0nmbRn6H8ndwLUQ==", sign)
}

func TestSigner_InvalidSecret(t *testing.T) {
	_, err := NewSigner("not base64!")
	require.Error(t, err)
}

func TestSigner_EmptySecret(t *testing.T) {
	signer, err := NewSigner("")
	require.NoError(t, err)

	_, err = signer.Sign("/0/private/Balance", "1", "nonce=1")
	require.Error(t, err)
}

func TestNonceSequence_Increasing(t *testing.T) {
	seq := &nonceSequence{sem: make(chan struct{}, 1)}
	var prev uint64
	for range 1000 {
		n, err := strconv.ParseUint(seq.next(), 10, 64)
		require.NoError(t, err)
		require.Greater(t, n, prev)
		prev = n
	}
}

func TestNonceFor_SharedPerKey(t *testing.T) {
	require.Same(t, nonceFor("key-a"), nonceFor("key-a"))
	require.NotSame(t, nonceFor("key-a"), nonceFor("key-b"))
}
