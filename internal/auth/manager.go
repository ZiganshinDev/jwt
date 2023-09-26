package auth

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Manager struct {
	signingKey string
}

type CustomClaims struct {
	jwt.StandardClaims
	GUID string `json:"guid"`
}

func NewManager(signingKey string) (*Manager, error) {
	if signingKey == "" {
		return nil, errors.New("empty signing key")
	}

	return &Manager{signingKey: signingKey}, nil
}

func (m *Manager) NewJWT(data string, ttl time.Duration) (string, error) {
	guid := uuid.New().String()

	claims := CustomClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ttl).Unix(),
			Subject:   data,
		},
		GUID: guid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	return token.SignedString([]byte(m.signingKey))
}

func (m *Manager) NewRefreshToken() (string, error) {
	const op = "auth.manager.NewRefreshToken"

	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return fmt.Sprintf("%x", b), nil
}

func (m *Manager) HashToken(token string) ([]byte, error) {
	const op = "auth.manager.HashToken"

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return []byte{}, fmt.Errorf("%s: %w", op, err)
	}

	return hashedToken, nil
}

func (m *Manager) CompareTokens(providedToken string, hashedToken []byte) bool {
	err := bcrypt.CompareHashAndPassword(hashedToken, []byte(providedToken))

	return err == nil
}
