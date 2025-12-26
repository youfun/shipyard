package utils

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

// Base58 alphabet (Bitcoin)
const b58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var b58Index = func() map[byte]int {
	m := make(map[byte]int, len(b58Alphabet))
	for i := 0; i < len(b58Alphabet); i++ {
		m[b58Alphabet[i]] = i
	}
	return m
}()

// encodeBase58 encodes raw bytes to a Base58 string
func encodeBase58(src []byte) string {
	// Count leading zeros
	zeros := 0
	for zeros < len(src) && src[zeros] == 0 {
		zeros++
	}
	// Copy source to mutable array
	input := make([]byte, len(src))
	copy(input, src)

	// Convert base-256 to base-58
	var encoded []byte
	startAt := zeros
	for startAt < len(input) {
		carry := 0
		for i := startAt; i < len(input); i++ {
			v := int(input[i]) + carry*256
			input[i] = byte(v / 58)
			carry = v % 58
		}
		encoded = append(encoded, b58Alphabet[carry])
		for startAt < len(input) && input[startAt] == 0 {
			startAt++
		}
	}

	// Add leading zeros as '1'
	for i := 0; i < zeros; i++ {
		encoded = append(encoded, '1')
	}

	// Reverse
	for i, j := 0, len(encoded)-1; i < j; i, j = i+1, j-1 {
		encoded[i], encoded[j] = encoded[j], encoded[i]
	}
	return string(encoded)
}

// decodeBase58 decodes a Base58 string to raw bytes
func decodeBase58(s string) ([]byte, error) {
	if s == "" {
		return []byte{}, nil
	}
	// Count leading '1's
	zeros := 0
	for zeros < len(s) && s[zeros] == '1' {
		zeros++
	}
	// Convert base-58 to base-256
	size := len(s)
	input := make([]byte, size)
	for i := 0; i < size; i++ {
		c := s[i]
		val, ok := b58Index[c]
		if !ok {
			return nil, errors.New("invalid base58 character")
		}
		carry := val
		for j := size - 1; j >= 0; j-- {
			carry += int(input[j]) * 58
			input[j] = byte(carry % 256)
			carry /= 256
		}
	}
	// Strip leading zeros from conversion output
	var out []byte
	// Find first non-zero
	firstNonZero := 0
	for firstNonZero < len(input) && input[firstNonZero] == 0 {
		firstNonZero++
	}
	out = input[firstNonZero:]
	// Add back leading zeros
	if zeros > 0 {
		out = append(make([]byte, zeros), out...)
	}
	return out, nil
}

// Prefix mapping per type
const (
	PrefixProject       = "prj_"
	PrefixApplication   = "app_"
	PrefixRelease       = "rel_"
	PrefixDeployment    = "dpl_"
	PrefixEnvVar        = "env_"
	PrefixRouting       = "rtg_"
	PrefixBuildTask     = "bld_"
	PrefixBuildArtifact = "bda_" // New prefix for BuildArtifact
	PrefixProviderAuth  = "pav_"
	PrefixAppToken      = "tok_"
	PrefixGitHubToken   = "ght_"
	PrefixUser          = "usr_"
	PrefixSSHHost       = "ssh_"
	PrefixAppInstance   = "inst_"
	PrefixDatabase      = "db_"
)

// EncodeFriendlyID returns prefix+base58(uuid_bytes)
func EncodeFriendlyID(prefix string, id uuid.UUID) string {
	return prefix + encodeBase58(id[:])
}

// DecodeFriendlyID strips expected prefix and decodes base58 to uuid
func DecodeFriendlyID(expectedPrefix string, friendly string) (uuid.UUID, error) {
	if !strings.HasPrefix(friendly, expectedPrefix) {
		return uuid.Nil, errors.New("invalid id prefix")
	}
	raw := strings.TrimPrefix(friendly, expectedPrefix)
	bytes, err := decodeBase58(raw)
	if err != nil {
		return uuid.Nil, err
	}
	if len(bytes) != 16 {
		return uuid.Nil, errors.New("invalid id length")
	}
	var id uuid.UUID
	copy(id[:], bytes)
	return id, nil
}

// PtrString utility
func PtrString(s string) *string { return &s }

// StringOrEmpty utility
func StringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ParseUUID parses a UUID string
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
