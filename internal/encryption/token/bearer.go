package token

import (
	"fmt"
	"strings"
	"time"

	"github.com/wtask/pwsrv/internal/core"
	"github.com/wtask/pwsrv/internal/core/middleware"
	"github.com/wtask/pwsrv/internal/encryption/hasher"
)

// AuthBearer - common internal interface to create and validate authorization tokens
type AuthBearer interface {
	middleware.TokenDiscoverer
	core.TokenProvider
}

type (
	payload struct {
		Issuer         string `json:""iss,omitempty`
		ExpirationTime int64  `json:"exp"`
		UserID         uint64 `json:"sub"`
	}

	bearer struct {
		ttl          time.Duration
		issuer       string
		timeProvider func() time.Time
		signer       hasher.StringHasher
	}

	bearerOption func(*bearer)
)

// NewMD5DigestBearer - initialize bearer with MD5-digest signing method.
func NewMD5DigestBearer(options ...bearerOption) AuthBearer {
	b := (&bearer{}).alter(options...)
	if b.signer == nil {
		b.signer = hasher.NewMD5DigestHasher("")
	}
	if b.timeProvider == nil {
		b.timeProvider = defaultTimeProvider
	}
	return b
}

func (b *bearer) alter(options ...bearerOption) *bearer {
	if b == nil {
		return nil
	}
	for _, o := range options {
		if o != nil {
			o(b)
		}
	}
	return b
}

// WithSignatureSecret - initialize bearer signature secret value.
func WithSignatureSecret(secret string) bearerOption {
	return func(b *bearer) {
		b.signer = hasher.NewMD5DigestHasher(secret)
	}
}

// WithTTL - initialize bearer with TTL value.
func WithTTL(ttl time.Duration) bearerOption {
	return func(b *bearer) {
		b.ttl = ttl
	}
}

// WithIssuer - initialize bearer with Issuer name.
func WithIssuer(issuer string) bearerOption {
	return func(b *bearer) {
		b.issuer = issuer
	}
}

func defaultTimeProvider() time.Time {
	return time.Now()
}

// withTimeProvider - initialize bearer with custom time provider.
func withTimeProvider(p func() time.Time) bearerOption {
	return func(b *bearer) {
		b.timeProvider = p
	}
}

// NewToken - return new token with given subject or empty string in case of error.
func (b *bearer) NewToken(userID uint64) string {
	if b == nil {
		return ""
	}
	t := b.timeProvider().UTC()
	p := payload{
		UserID:         userID,
		Issuer:         b.issuer,
		ExpirationTime: t.Add(b.ttl).Unix(),
	}
	b64 := encodeJSONB64(&p)
	if b64 == "" {
		return ""
	}
	sig := b.signer.Hash(b64)
	if sig == "" {
		return ""
	}
	return fmt.Sprintf("%s.%s", b64, sig)
}

func (b *bearer) assertToken(token string) *payload {
	if b == nil || token == "" {
		return nil
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 ||
		// if 2 - only second part may be empty
		parts[1] == "" {
		return nil
	}
	if parts[1] != b.signer.Hash(parts[0]) {
		return nil
	}
	p := &payload{}
	if !decodeJSONB64(parts[0], p) {
		return nil
	}
	return p
}

// DiscoverUserID - middleware.AuthBearer implementation.
func (b *bearer) DiscoverUserID(token string) (uint64, bool) {
	if b == nil {
		return 0, false
	}
	p := b.assertToken(token)
	if p == nil {
		return 0, false
	}
	ok := b.timeProvider().UTC().Unix() < p.ExpirationTime &&
		b.issuer == p.Issuer
	return p.UserID, ok
}
