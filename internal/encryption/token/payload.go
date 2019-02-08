// Package token
package token

// Payload - unsigned, base64-url-encoded token section.
type Payload struct {
	Issuer         string `json:""iss,omitempty`
	ExpirationTime int64  `json:"exp"`
	UserID         uint64 `json:"sub"`
	Email          string `json:"email,omitempty"`
}

type payloadClaim = func(*Payload)

func WithIssuer(iss string) payloadClaim {
	return func(p *Payload) {
		p.Issuer = iss
	}
}

func WithExpiration(expireAfter int64) payloadClaim {
	return func(p *Payload) {
		p.ExpirationTime = expireAfter
	}
}

func WithSubject(userID uint64) payloadClaim {
	return func(p *Payload) {
		p.UserID = userID
	}
}

func WithEmail(email string) payloadClaim {
	return func(p *Payload) {
		p.Email = email
	}
}

func NewPayload(claims ...payloadClaim) *Payload {
	p := &Payload{}
	UpdatePayload(p, claims...)
	return p
}

func UpdatePayload(p *Payload, claims ...payloadClaim) {
	for _, c := range claims {
		c(p)
	}
}
