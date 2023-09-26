package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	signingKey := "qwerty"

	_, err := New(signingKey)
	require.NoError(t, err)
}

func TestNewManagerError(t *testing.T) {
	signingKey := ""

	_, err := New(signingKey)
	require.Error(t, err)
}

func TestNewJWT(t *testing.T) {
	m := Manager{}
	data := "data"
	ttl := time.Duration(time.Duration.Hours(5))

	jwt, err := m.NewJWT(data, ttl)
	require.NoError(t, err)
	require.NotEmpty(t, jwt)
}

func TestHashToken(t *testing.T) {
	m := Manager{}
	token := "4373fbac63c46617971af8e9127cc69dcb8981e3f9b285b08f0b76c6501f7256"

	_, err := m.HashToken(token)
	require.NoError(t, err)
}

func TestCompareTokens(t *testing.T) {
	m := Manager{}
	providedToken := "4373fbac63c46617971af8e9127cc69dcb8981e3f9b285b08f0b76c6501f7256"

	hashedToken, _ := m.HashToken(providedToken)

	ok := m.CompareTokens(providedToken, hashedToken)
	require.True(t, ok)
}

func TestCompareTokensError(t *testing.T) {
	m := Manager{}
	providedToken := "4373fbac63c46617971af8e9127cc69dcb8981e3f9b285b08f0b76c6501f7256"

	hashedToken := []byte("")

	ok := m.CompareTokens(providedToken, hashedToken)
	require.False(t, ok)
}
