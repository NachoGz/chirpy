package auth

import (
	"testing"
	"time"
	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	// Generate a test user ID
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := time.Second * 2

	// Create a JWT
	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("failed to create JWT: %v", err)
	}

	// Validate the JWT
	parsedUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("failed to validate JWT: %v", err)
	}

	if parsedUserID != userID {
		t.Errorf("expected userID %s, got %s", userID, parsedUserID)
	}
}

func TestExpiredJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"

	// Create a token that expires immediately
	token, err := MakeJWT(userID, tokenSecret, time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create JWT: %v", err)
	}

	// Wait for token to expire
	time.Sleep(time.Millisecond * 2)

	_, err = ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Fatal("expected error for expired JWT, got nil")
	}
}

func TestInvalidSecretJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "correct-secret"

	// Create a valid token
	token, err := MakeJWT(userID, tokenSecret, time.Minute)
	if err != nil {
		t.Fatalf("failed to create JWT: %v", err)
	}

	// Attempt validation with the wrong secret
	_, err = ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for invalid secret, got nil")
	}
}

func TestHashPassword(t *testing.T) {
	password := "securepassword"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// Ensure the hashed password is not empty
	if hashedPassword == "" {
		t.Fatal("hashed password is empty")
	}

	// Ensure the hashed password is not the same as the input password
	if hashedPassword == password {
		t.Fatal("hashed password should not match the original password")
	}

	// Verify that the hash can be checked successfully
	err = CheckPasswordHash(password, hashedPassword)
	if err != nil {
		t.Fatalf("failed to verify hashed password: %v", err)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "mypassword"

	// Manually hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate bcrypt hash: %v", err)
	}

	// Test valid password
	err = CheckPasswordHash(password, string(hashedPassword))
	if err != nil {
		t.Fatalf("expected password to be valid, got error: %v", err)
	}

	// Test invalid password
	err = CheckPasswordHash("wrongpassword", string(hashedPassword))
	if err == nil {
		t.Fatal("expected error for incorrect password, but got none")
	}
}
