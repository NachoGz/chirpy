package auth

import (
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
	"errors"
	"strings"
	"net/http"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func HashPassword(password string) (string, error) {
	hashed_passwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed_passwd), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims {
		Issuer:		"chirpy",
		IssuedAt:	jwt.NewNumericDate(time.Now()),
		ExpiresAt:	jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:	userID.String(),
	})

	signed_token, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signed_token, nil
}

// ValidateJWT parses and validates a JWT, returning the user ID if valid.
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	// Check if the token is valid and not expired
	if !token.Valid {
		return uuid.UUID{}, errors.New("invalid token")
	}

	// Parse the user ID from the Subject field
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, errors.New("invalid user ID in token")
	}

	return userID, nil
}


func GetBearerToken(headers http.Header) (string, error) {
	auth_header := headers.Get("Authorization")
	if auth_header == "" {
		return "", errors.New("The TOKEN_STRING doesn't exist")
	}

	return strings.TrimSpace(strings.TrimPrefix(auth_header, "Bearer ")), nil
}


func MakeRefreshToken() (string, error) {
	// Create a 32-byte slice to hold the random data
	random_data := make([]byte, 32)


	// Read 32 random bytes and store them in random_data
	_, err := rand.Read(random_data)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert random bytes to hex
	return hex.EncodeToString(random_data), nil
}