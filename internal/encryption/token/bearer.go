package token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/wtask/pwsrv/internal/encryption/hasher"
)

// AuthBearer - common internal interface to create and validate authorization tokens
type AuthBearer interface {
	CreateToken(claims ...payloadClaim) (token string)
	GetPayload(token string) (p *Payload, ok bool)
}

type bearer struct {
	ttl    time.Duration
	signer hasher.StringHasher
}

// NewMD5DigestBearer - initialize bearer with MD5-digest signing method.
func NewMD5DigestBearer(secret string, ttl time.Duration) AuthBearer {
	return &bearer{
		ttl:    ttl,
		signer: hasher.NewMD5DigestHasher(secret),
	}
}

// CreateToken - returns encrypted token corresponding passed claims
// or empty string in case of error.
func (b *bearer) CreateToken(claims ...payloadClaim) string {
	if b == nil {
		return ""
	}
	p := NewPayload(WithExpiration(time.Now().UTC().Add(b.ttl).Unix()))
	UpdatePayload(p, claims...)
	return b.encodeToken(p)
}

// GetPayload - return token payload and token validation result.
// Unknown, expired tokens, tokens with zero user ID are not valid.
func (b *bearer) GetPayload(token string) (p *Payload, ok bool) {
	if b == nil {
		return nil, false
	}
	p = b.decodeToken(token)
	ok = p != nil &&
		p.UserID > 0 &&
		time.Now().UTC().Unix() <= p.ExpirationTime
	return p, ok
}

// encodeToken - encodes payload data and returns non-empty token.
func (b *bearer) encodeToken(p *Payload) string {
	if b == nil || b.signer == nil {
		return ""
	}
	p64, ok := encodeJSONB64(p)
	if !ok {
		return ""
	}
	s := b.signer.Hash(p64)
	if s == "" {
		return ""
	}
	return fmt.Sprintf("%s.%s", p64, s)
}

// decodeToken - decodes payload from valid token.
// Returns nil if can't decode.
func (b *bearer) decodeToken(token string) *Payload {
	if b == nil || token == "" {
		return nil
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 ||
		parts[0] == "" {
		return nil
	}
	p := &Payload{}
	if !decodeJSONB64(parts[0], p) {
		return nil
	}
	if b.encodeToken(p) != token {
		return nil
	}
	return p
}

func encodeJSONB64(v interface{}) (string, bool) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", false
	}
	return base64.RawURLEncoding.EncodeToString(bytes), true
}

func decodeJSONB64(src string, v interface{}) bool {
	bytes, err := base64.RawURLEncoding.DecodeString(src)
	if err != nil {
		return false
	}
	if err = json.Unmarshal(bytes, v); err != nil {
		return false
	}
	return true
}
