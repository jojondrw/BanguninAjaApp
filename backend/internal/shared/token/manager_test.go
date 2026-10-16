package token

import (
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
)

const (
	testSecret = "rahasia-uji-yang-cukup-panjang-untuk-hs256"
	testIssuer = "banguninaja-uji"
)

var urlSafeRefresh = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

func newTestManager(secret, issuer string, accessTTL time.Duration) *Manager {
	return NewManager(config.Token{
		AccessSecret: secret,
		AccessTTL:    accessTTL,
		RefreshTTL:   7 * 24 * time.Hour,
		Issuer:       issuer,
	})
}

func signedWith(t *testing.T, method jwt.SigningMethod, key any, claims jwt.RegisteredClaims) string {
	t.Helper()
	value, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}
	return value
}

func validClaims(subject string) jwt.RegisteredClaims {
	now := time.Now()
	return jwt.RegisteredClaims{
		Subject:   subject,
		Issuer:    testIssuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
	}
}

func TestIssuedAccessTokenParsesBackToTheSameUser(t *testing.T) {
	manager := newTestManager(testSecret, testIssuer, 15*time.Minute)
	userID := uuid.New()

	access, err := manager.IssueAccess(userID)
	if err != nil {
		t.Fatalf("issue access: %v", err)
	}
	if remaining := time.Until(access.ExpiresAt); remaining <= 14*time.Minute || remaining > 15*time.Minute {
		t.Fatalf("access token expires in %v, want about 15 minutes", remaining)
	}

	parsed, err := manager.ParseAccess(access.Value)
	if err != nil {
		t.Fatalf("parse access: %v", err)
	}
	if parsed != userID {
		t.Fatalf("parsed user %s, want %s", parsed, userID)
	}
}

func TestParseAccessRejectsTokensThatMustNotBeTrusted(t *testing.T) {
	manager := newTestManager(testSecret, testIssuer, 15*time.Minute)
	userID := uuid.New()

	expired, err := newTestManager(testSecret, testIssuer, -time.Minute).IssueAccess(userID)
	if err != nil {
		t.Fatalf("issue expired token: %v", err)
	}
	otherSecret, err := newTestManager("rahasia-lain-yang-juga-panjang-sekali", testIssuer, time.Minute).IssueAccess(userID)
	if err != nil {
		t.Fatalf("issue token with another secret: %v", err)
	}
	otherIssuer, err := newTestManager(testSecret, "penerbit-lain", time.Minute).IssueAccess(userID)
	if err != nil {
		t.Fatalf("issue token from another issuer: %v", err)
	}

	withoutExpiry := validClaims(userID.String())
	withoutExpiry.ExpiresAt = nil

	forUser := signedWith(t, jwt.SigningMethodHS256, []byte(testSecret), validClaims(userID.String()))
	forOther := signedWith(t, jwt.SigningMethodHS256, []byte(testSecret), validClaims(uuid.NewString()))
	ownParts := strings.Split(forUser, ".")
	otherParts := strings.Split(forOther, ".")
	swappedPayload := ownParts[0] + "." + otherParts[1] + "." + ownParts[2]

	cases := map[string]string{
		"expired":                  expired.Value,
		"signed with other secret": otherSecret.Value,
		"from another issuer":      otherIssuer.Value,
		"signed with HS512":        signedWith(t, jwt.SigningMethodHS512, []byte(testSecret), validClaims(userID.String())),
		"unsigned alg none":        signedWith(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, validClaims(userID.String())),
		"without expiry":           signedWith(t, jwt.SigningMethodHS256, []byte(testSecret), withoutExpiry),
		"subject is not a uuid":    signedWith(t, jwt.SigningMethodHS256, []byte(testSecret), validClaims("admin")),
		"payload swapped":          swappedPayload,
		"not a jwt":                "bukan-token",
		"empty":                    "",
	}

	for name, value := range cases {
		t.Run(name, func(t *testing.T) {
			parsed, err := manager.ParseAccess(value)
			if !errors.Is(err, ErrInvalidAccessToken) {
				t.Fatalf("err = %v, want ErrInvalidAccessToken", err)
			}
			if parsed != uuid.Nil {
				t.Fatalf("parsed user %s from an untrusted token", parsed)
			}
		})
	}
}

func TestIssueRefreshReturnsARandomURLSafeValueAndItsHash(t *testing.T) {
	manager := newTestManager(testSecret, testIssuer, time.Minute)

	first, err := manager.IssueRefresh()
	if err != nil {
		t.Fatalf("issue refresh: %v", err)
	}
	second, err := manager.IssueRefresh()
	if err != nil {
		t.Fatalf("issue refresh: %v", err)
	}

	if !urlSafeRefresh.MatchString(first.Value) {
		t.Fatalf("refresh value %q is not 32 random bytes in URL-safe base64", first.Value)
	}
	if first.Value == second.Value || first.Hash == second.Hash {
		t.Fatal("two refresh tokens are identical")
	}
	if first.Hash != HashRefresh(first.Value) || first.Hash == first.Value {
		t.Fatal("refresh hash does not match HashRefresh of its value")
	}
	if remaining := time.Until(first.ExpiresAt); remaining <= 7*24*time.Hour-time.Minute || remaining > 7*24*time.Hour {
		t.Fatalf("refresh token expires in %v, want about 7 days", remaining)
	}
}

func TestHashRefreshIsHexSHA256(t *testing.T) {
	const sha256OfABC = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"

	if got := HashRefresh("abc"); got != sha256OfABC {
		t.Fatalf("HashRefresh(abc) = %s, want %s", got, sha256OfABC)
	}
	if HashRefresh("abc") != HashRefresh("abc") {
		t.Fatal("HashRefresh is not deterministic")
	}
}
