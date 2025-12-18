package helpers

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Global secret key from environment variable
var secretKey = os.Getenv("JWT_SECRET_KEY")

// Claims struct defines the structure of the JWT payload
type Claims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// CreateTokenFunc is an overridable function (default implementation below) so tests can mock it.
var CreateTokenFunc = func(userID uuid.UUID, email string) (string, error) {
	// Set the expiration time for the token (e.g., 24 hours from now)
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		ID:    userID.String(),
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create the token using the claims and HMAC signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// CreateToken calls the overridable GenerateTokenFunc.
func CreateToken(userID uuid.UUID, email string) (string, error) {
	return CreateTokenFunc(userID, email)
}

// VerifyToken verifies the JWT token string and returns the claims (payload).
func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	// 1. Parse the token
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify that the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, errors.New("invalid token: " + err.Error())
	}

	// 2. Check if the token is valid and get claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token or claims")
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
