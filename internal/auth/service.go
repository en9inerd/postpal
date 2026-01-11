package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	defaultMemory      = 64 * 1024
	defaultTime        = 3
	defaultParallelism = 2
	defaultKeyLength   = 32
)

type argon2Params struct {
	memory      uint32
	time        uint32
	parallelism uint8
	keyLength   uint32
}

type Service struct {
	passwordHashEncoded string
	sessionSecret       []byte
	sessionMaxAge       time.Duration
}

func NewService(passwordHashEncoded, sessionSecret string, maxAgeSeconds int) (*Service, error) {
	if passwordHashEncoded == "" {
		return nil, errors.New("password hash is required (set AUTH_PASSWORD_HASH environment variable)")
	}
	if !isArgon2idHash(passwordHashEncoded) {
		return nil, fmt.Errorf("invalid Argon2id hash format: hash must start with '$argon2id$' (generate one using: go run scripts/generate-password-hash.go \"your-password\")")
	}

	secretBytes, err := base64.StdEncoding.DecodeString(sessionSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid session secret: %w", err)
	}

	if len(secretBytes) < 32 {
		return nil, errors.New("session secret must be at least 32 bytes")
	}

	return &Service{
		passwordHashEncoded: passwordHashEncoded,
		sessionSecret:       secretBytes,
		sessionMaxAge:       time.Duration(maxAgeSeconds) * time.Second,
	}, nil
}

func (s *Service) VerifyPassword(password string) error {
	match, err := comparePasswordAndHash(password, s.passwordHashEncoded)
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}
	if !match {
		return errors.New("invalid password")
	}
	return nil
}

func (s *Service) GenerateSessionToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	h := hmac.New(sha256.New, s.sessionSecret)
	h.Write([]byte(token))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	return token + "." + signature, nil
}

func (s *Service) ValidateSessionToken(signedToken string) (bool, error) {
	parts := strings.Split(signedToken, ".")
	if len(parts) != 2 {
		return false, errors.New("invalid token format")
	}

	token, signature := parts[0], parts[1]
	h := hmac.New(sha256.New, s.sessionSecret)
	h.Write([]byte(token))
	expectedSig := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return false, errors.New("invalid token signature")
	}

	return true, nil
}

func (s *Service) GetSessionMaxAge() time.Duration {
	return s.sessionMaxAge
}

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		defaultTime,
		defaultMemory,
		defaultParallelism,
		defaultKeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		defaultMemory,
		defaultTime,
		defaultParallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

func comparePasswordAndHash(password, encodedHash string) (bool, error) {
	params, salt, hash, err := decodeArgon2idHash(encodedHash)
	if err != nil {
		return false, err
	}

	derivedKey := argon2.IDKey(
		[]byte(password),
		salt,
		params.time,
		params.memory,
		params.parallelism,
		params.keyLength,
	)

	return subtle.ConstantTimeCompare(derivedKey, hash) == 1, nil
}

func decodeArgon2idHash(encodedHash string) (*argon2Params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, errors.New("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, errors.New("not an Argon2id hash")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return nil, nil, nil, errors.New("incompatible version")
	}

	var params argon2Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.memory, &params.time, &params.parallelism); err != nil {
		return nil, nil, nil, fmt.Errorf("invalid parameters: %w", err)
	}
	params.keyLength = defaultKeyLength

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid hash: %w", err)
	}

	return &params, salt, hash, nil
}

func isArgon2idHash(hash string) bool {
	return strings.HasPrefix(hash, "$argon2id$")
}
