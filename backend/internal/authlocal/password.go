package authlocal

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Password hashing mirrors the IAM service's internal/auth/password package:
// argon2id with the standard PHC-style encoded string
// ("$argon2id$v=19$m=...,t=...,p=...$salt$hash"), so parameters travel with the
// hash and can be tuned without a schema change.

// hashParams are the argon2id cost parameters. Defaults follow OWASP guidance
// (64 MiB, 1 iteration, parallelism 4) and are encoded into each hash.
type hashParams struct {
	memory      uint32 // KiB
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var defaultHashParams = hashParams{
	memory:      64 * 1024,
	iterations:  1,
	parallelism: 4,
	saltLength:  16,
	keyLength:   32,
}

// errMismatch is returned by verifyPassword when the password does not match.
var errMismatch = errors.New("password mismatch")

// errInvalidHash is returned when an encoded hash is malformed or uses an
// unsupported algorithm/version.
var errInvalidHash = errors.New("invalid password hash")

// hashPassword derives an argon2id hash of plaintext and returns the
// PHC-encoded string.
func hashPassword(plaintext string) (string, error) {
	p := defaultHashParams
	salt := make([]byte, p.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read salt: %w", err)
	}
	key := argon2.IDKey([]byte(plaintext), salt, p.iterations, p.memory, p.parallelism, p.keyLength)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.memory, p.iterations, p.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// verifyPassword reports whether plaintext matches encoded. It returns
// errMismatch on a valid-but-non-matching password and errInvalidHash on a
// malformed encoding. The comparison is constant-time.
func verifyPassword(plaintext, encoded string) error {
	p, salt, key, err := decodeHash(encoded)
	if err != nil {
		return err
	}
	candidate := argon2.IDKey([]byte(plaintext), salt, p.iterations, p.memory, p.parallelism, uint32(len(key)))
	if subtle.ConstantTimeCompare(key, candidate) != 1 {
		return errMismatch
	}
	return nil
}

func decodeHash(encoded string) (hashParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	// "", "argon2id", "v=19", "m=...,t=...,p=...", salt, hash
	if len(parts) != 6 || parts[1] != "argon2id" {
		return hashParams{}, nil, nil, errInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return hashParams{}, nil, nil, errInvalidHash
	}

	var p hashParams
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism); err != nil {
		return hashParams{}, nil, nil, errInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return hashParams{}, nil, nil, errInvalidHash
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return hashParams{}, nil, nil, errInvalidHash
	}
	return p, salt, key, nil
}
