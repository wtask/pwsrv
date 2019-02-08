// Package hasher contains types and methods for hashing values.
package hasher

import (
	"crypto/md5"
	"fmt"
)

// StringHasher - common internal interface to hash any string
type StringHasher interface {
	Hash(string) string
}

// SecureHasher - provides string hashing.
type secureHasher struct {
	secret string
	encode func(string) string
}

// Hash - hashes passed string.
func (h *secureHasher) Hash(s string) string {
	if h == nil ||
		h.encode == nil ||
		(h.secret == "" && s == "") {
		return ""
	}
	return h.encode(fmt.Sprintf("%[1]s%[2]s%[1]s", h.secret, s))
}

// NewMD5DigestHasher - implements MD5 digest hashing (in uppercase).
func NewMD5DigestHasher(secret string) StringHasher {
	return &secureHasher{
		secret: secret,
		encode: func(s string) string {
			return fmt.Sprintf("%X", md5.Sum([]byte(s)))
		},
	}
}
