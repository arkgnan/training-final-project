package helpers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHashAndCheckPassword(t *testing.T) {
	t.Parallel()

	password := "MyStr0ngP@ss!"
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash, "hash should not be empty")

	// Correct password should validate
	assert.True(t, CheckPasswordHash(password, hash), "expected password & hash to match")

	// Wrong password should not validate
	assert.False(t, CheckPasswordHash("wrong-password", hash), "expected wrong password to not match hash")
}

func TestCreateAndVerifyToken(t *testing.T) {
	t.Parallel()

	// Set test secret key (same-package test can access unexported var)
	secretKey = "unit-test-secret"

	userID := uuid.New()
	email := "tester@example.com"

	token, err := CreateToken(userID, email)
	assert.NoError(t, err)
	assert.NotEmpty(t, token, "token should be created")

	claims, err := VerifyToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	// jwt.MapClaims stores values as interface{}; assert values equal
	assert.Equal(t, userID.String(), claims["id"], "claim id should match")
	assert.Equal(t, email, claims["email"], "claim email should match")
}

func TestVerifyInvalidToken(t *testing.T) {
	t.Parallel()

	// Ensure secret key is set to something predictable
	secretKey = "unit-test-secret"

	_, err := VerifyToken("this.is.not.a.valid.token")
	assert.Error(t, err, "invalid token should return an error")
}
