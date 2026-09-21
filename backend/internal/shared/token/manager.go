package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
)

const refreshTokenBytes = 32

var ErrInvalidAccessToken = errors.New("access token is not valid")

type Manager struct {
	accessSecret []byte
	accessTTL    time.Duration
	refreshTTL   time.Duration
	issuer       string
}

type Access struct {
	Value     string
	ExpiresAt time.Time
}

type Refresh struct {
	Value     string
	Hash      string
	ExpiresAt time.Time
}

func NewManager(cfg config.Token) *Manager {
	return &Manager{
		accessSecret: []byte(cfg.AccessSecret),
		accessTTL:    cfg.AccessTTL,
		refreshTTL:   cfg.RefreshTTL,
		issuer:       cfg.Issuer,
	}
}

func (m *Manager) IssueAccess(userID uuid.UUID) (Access, error) {
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(m.accessTTL)

	claims := jwt.RegisteredClaims{
		ID:        uuid.NewString(),
		Subject:   userID.String(),
		Issuer:    m.issuer,
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	value, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.accessSecret)
	if err != nil {
		return Access{}, fmt.Errorf("sign access token: %w", err)
	}

	return Access{Value: value, ExpiresAt: expiresAt}, nil
}

func (m *Manager) ParseAccess(value string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}

	parsed, err := jwt.ParseWithClaims(value, claims, m.keyFunc,
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !parsed.Valid {
		return uuid.Nil, ErrInvalidAccessToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidAccessToken
	}

	return userID, nil
}

func (m *Manager) IssueRefresh() (Refresh, error) {
	buffer := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buffer); err != nil {
		return Refresh{}, fmt.Errorf("generate refresh token: %w", err)
	}

	value := base64.RawURLEncoding.EncodeToString(buffer)

	return Refresh{
		Value:     value,
		Hash:      HashRefresh(value),
		ExpiresAt: time.Now().Add(m.refreshTTL),
	}, nil
}

func HashRefresh(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (m *Manager) keyFunc(*jwt.Token) (any, error) {
	return m.accessSecret, nil
}
