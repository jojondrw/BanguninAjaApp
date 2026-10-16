package security

import (
	"strings"
	"testing"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPasswordStoresABcryptHashThatOnlyMatchesTheSamePassword(t *testing.T) {
	hashed, err := HashPassword("kata-sandi-benar")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if hashed == "kata-sandi-benar" || !strings.HasPrefix(hashed, "$2a$") {
		t.Fatalf("hash %q is not a bcrypt hash", hashed)
	}
	cost, err := bcrypt.Cost([]byte(hashed))
	if err != nil || cost != bcrypt.DefaultCost {
		t.Fatalf("cost = %d (%v), want %d", cost, err, bcrypt.DefaultCost)
	}
	if !PasswordMatches(hashed, "kata-sandi-benar") {
		t.Fatal("the original password does not match its hash")
	}
	for _, wrong := range []string{"kata-sandi-salah", "Kata-sandi-benar", "", "kata-sandi-benar "} {
		if PasswordMatches(hashed, wrong) {
			t.Fatalf("password %q matched a hash of a different password", wrong)
		}
	}
}

func TestHashPasswordSaltsEveryHash(t *testing.T) {
	first, err := HashPassword("sama-persis")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	second, err := HashPassword("sama-persis")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if first == second {
		t.Fatal("two hashes of the same password are identical, so the salt is missing")
	}
	if !PasswordMatches(first, "sama-persis") || !PasswordMatches(second, "sama-persis") {
		t.Fatal("a salted hash no longer matches its password")
	}
}

func TestPasswordMatchesRejectsAStoredValueThatIsNotABcryptHash(t *testing.T) {
	for _, stored := range []string{"", "kata-sandi-polos", "$2a$10$terlalu-pendek"} {
		if PasswordMatches(stored, "kata-sandi-polos") {
			t.Fatalf("stored value %q was accepted as a hash", stored)
		}
	}
}

func TestHashPasswordRejectsPasswordsLongerThanSeventyTwoBytes(t *testing.T) {
	if _, err := HashPassword(strings.Repeat("a", 72)); err != nil {
		t.Fatalf("72 byte password: %v", err)
	}
	if _, err := HashPassword(strings.Repeat("a", 73)); err == nil {
		t.Fatal("73 byte password was hashed, but bcrypt only reads the first 72 bytes")
	}

	accented := strings.Repeat("é", 72)
	if utf8.RuneCountInString(accented) != 72 || len(accented) != 144 {
		t.Fatalf("test input has %d characters and %d bytes", utf8.RuneCountInString(accented), len(accented))
	}
	if _, err := HashPassword(accented); err == nil {
		t.Fatal("72 accented characters (144 bytes) were hashed")
	}
}

func TestPasswordTooLongCountsBytesNotCharacters(t *testing.T) {
	if PasswordTooLong(strings.Repeat("a", MaxPasswordBytes)) {
		t.Fatal("a 72 byte password was reported as too long")
	}
	if !PasswordTooLong(strings.Repeat("a", MaxPasswordBytes+1)) {
		t.Fatal("a 73 byte password was accepted")
	}
	if !PasswordTooLong(strings.Repeat("é", 40)) {
		t.Fatal("40 accented characters are 80 bytes and must be too long")
	}
}
