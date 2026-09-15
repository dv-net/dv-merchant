package tools_test

import (
	"strings"
	"testing"

	"github.com/dv-net/dv-merchant/internal/tools"

	"github.com/stretchr/testify/require"
)

func Test_HashPassword(t *testing.T) {
	password := "super-secret-password"

	hash, err := tools.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	require.NotEqual(t, password, hash)
	require.True(t, strings.HasPrefix(hash, "$2a$"))
}

func Test_HashPassword_DifferentSaltEachTime(t *testing.T) {
	password := "super-secret-password"

	hash1, err := tools.HashPassword(password)
	require.NoError(t, err)

	hash2, err := tools.HashPassword(password)
	require.NoError(t, err)

	require.NotEqual(t, hash1, hash2)
}

func Test_CheckPasswordHash_Success(t *testing.T) {
	password := "super-secret-password"

	hash, err := tools.HashPassword(password)
	require.NoError(t, err)

	require.True(t, tools.CheckPasswordHash(password, hash))
}

func Test_CheckPasswordHash_WrongPassword(t *testing.T) {
	hash, err := tools.HashPassword("super-secret-password")
	require.NoError(t, err)

	require.False(t, tools.CheckPasswordHash("wrong-password", hash))
}

func Test_CheckPasswordHash_InvalidHash(t *testing.T) {
	require.False(t, tools.CheckPasswordHash("super-secret-password", "not-a-bcrypt-hash"))
}

func Test_CheckPasswordHash_EmptyPassword(t *testing.T) {
	hash, err := tools.HashPassword("")
	require.NoError(t, err)

	require.True(t, tools.CheckPasswordHash("", hash))
	require.False(t, tools.CheckPasswordHash("not-empty", hash))
}
