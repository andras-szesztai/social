package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MockAuth struct{}

func NewMockAuth() Authenticator {
	return &MockAuth{}
}

var testClaims = jwt.MapClaims{
	"aud": "test-aud",
	"iss": "test-iss",
	"sub": 1,
	"exp": time.Now().Add(time.Hour * 24).Unix(),
	"iat": time.Now().Unix(),
	"nbf": time.Now().Unix(),
}

func (m *MockAuth) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims)
	tokenString, _ := token.SignedString([]byte("test-key"))
	return tokenString, nil
}

func (m *MockAuth) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-key"), nil
	})
}
