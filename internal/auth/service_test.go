package auth

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("test-password")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("hash should start with $argon2id$, got: %s", hash)
	}

	if len(hash) < 50 {
		t.Errorf("hash seems too short: %d chars", len(hash))
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "test-password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	sessionSecret := base64.StdEncoding.EncodeToString([]byte("test-secret-that-is-exactly-32-bytes-long"))
	service, err := NewService(hash, sessionSecret, 3600)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	if err := service.VerifyPassword(password); err != nil {
		t.Errorf("VerifyPassword failed for correct password: %v", err)
	}

	if err := service.VerifyPassword("wrong-password"); err == nil {
		t.Error("VerifyPassword should fail for wrong password")
	}
}

func TestSessionToken(t *testing.T) {
	hash, _ := HashPassword("test")
	sessionSecret := base64.StdEncoding.EncodeToString([]byte("test-secret-that-is-exactly-32-bytes-long"))

	service, err := NewService(hash, sessionSecret, 3600)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	token, err := service.GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}

	if token == "" {
		t.Error("token should not be empty")
	}

	if !strings.Contains(token, ".") {
		t.Error("token should contain a dot separator")
	}

	valid, err := service.ValidateSessionToken(token)
	if err != nil {
		t.Fatalf("ValidateSessionToken failed: %v", err)
	}
	if !valid {
		t.Error("ValidateSessionToken should return true for valid token")
	}

	valid, err = service.ValidateSessionToken("invalid.token")
	if err == nil || valid {
		t.Error("ValidateSessionToken should fail for invalid token")
	}

	valid, err = service.ValidateSessionToken("not-a-valid-token")
	if err == nil || valid {
		t.Error("ValidateSessionToken should fail for malformed token")
	}
}

func TestNewServiceValidation(t *testing.T) {
	hash, _ := HashPassword("test")
	validSecret := base64.StdEncoding.EncodeToString([]byte("test-secret-that-is-exactly-32-bytes-long"))

	t.Run("invalid hash format", func(t *testing.T) {
		_, err := NewService("invalid-hash", validSecret, 3600)
		if err == nil {
			t.Error("NewService should fail for invalid hash")
		}
	})

	t.Run("invalid session secret", func(t *testing.T) {
		_, err := NewService(hash, "not-base64", 3600)
		if err == nil {
			t.Error("NewService should fail for invalid base64 secret")
		}
	})

	t.Run("short session secret", func(t *testing.T) {
		shortSecret := base64.StdEncoding.EncodeToString([]byte("short"))
		_, err := NewService(hash, shortSecret, 3600)
		if err == nil {
			t.Error("NewService should fail for secret < 32 bytes")
		}
	})

	t.Run("valid service", func(t *testing.T) {
		_, err := NewService(hash, validSecret, 3600)
		if err != nil {
			t.Errorf("NewService should succeed for valid inputs: %v", err)
		}
	})
}

func TestGetSessionMaxAge(t *testing.T) {
	hash, _ := HashPassword("test")
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-that-is-exactly-32-bytes-long"))

	service, err := NewService(hash, secret, 7200)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	if service.GetSessionMaxAge().Seconds() != 7200 {
		t.Errorf("expected max age 7200s, got %v", service.GetSessionMaxAge())
	}
}
