package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndValidateToken_Success(t *testing.T) {
	secret := "test-secret-key"
	token, err := GenerateToken("user-123", "user@example.com", secret, 24)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := ValidateToken(token, secret)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.Equal(t, "fortis-backend", claims.Issuer)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	secret := "test-secret-key"
	token, err := GenerateToken("user-123", "user@example.com", secret, 24)
	require.NoError(t, err)

	_, err = ValidateToken(token, "wrong-secret")
	require.Error(t, err)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	_, err := ValidateToken("invalid-token-string", "secret")
	require.Error(t, err)
}

func TestGenerateToken_EmptyUserID(t *testing.T) {
	secret := "test-secret"
	token, err := GenerateToken("", "user@example.com", secret, 24)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := ValidateToken(token, secret)
	require.NoError(t, err)
	assert.Empty(t, claims.UserID)
}
