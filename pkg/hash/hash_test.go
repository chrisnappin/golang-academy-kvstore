package hash_test

import (
	"fmt"
	"store/pkg/hash"
	"testing"

	"golang.org/x/crypto/argon2"
)

func TestGeneratedSameIsVerified(t *testing.T) {
	password := "password1234"
	encodedHash, err := hash.GenerateHash(password)
	if err != nil {
		t.Fatal("Error generating hash: ", err)
	}

	verified, err := hash.VerifyAgainstHash(password, encodedHash)
	if err != nil {
		t.Fatal("Error verifying hash: ", err)
	}

	if !verified {
		t.Fatal("Verifying same password against hash should have succeeded")
	}
}

func TestGeneratedDifferentIsNotVerified(t *testing.T) {
	password1 := "password1234"
	password2 := "wibble567"

	encodedHash, err := hash.GenerateHash(password1)
	if err != nil {
		t.Fatal("Error generating hash: ", err)
	}

	verified, err := hash.VerifyAgainstHash(password2, encodedHash)
	if err != nil {
		t.Fatal("Error verifying hash: ", err)
	}

	if verified {
		t.Fatal("Verifying different password against hash should have failed")
	}
}

func TestInvalidEncodedHash(t *testing.T) {
	_, err := hash.VerifyAgainstHash("password", "invalidEncodedHash")
	if err == nil {
		t.Fatal("Invalid encoded hash should have been rejected")
	}
}

func TestIncompatibleVersionEncodedHash(t *testing.T) {
	incompatibleEncodedHash := fmt.Sprintf("$argon2id$v=%d$m=1,t=2,p=3$YWE$Yg", argon2.Version+42)
	_, err := hash.VerifyAgainstHash("password", incompatibleEncodedHash)
	if err == nil {
		t.Fatal("Invalid argon2 version should have been rejected")
	}
}

func TestCompatibleVersionIsUnverified(t *testing.T) {
	compatibleEncodedHash := fmt.Sprintf("$argon2id$v=%d$m=1,t=2,p=3$YWE$Yg", argon2.Version)
	_, err := hash.VerifyAgainstHash("password", compatibleEncodedHash)
	if err != nil {
		t.Fatal("Compatible format hash should have been accepted: ", err)
	}
}
