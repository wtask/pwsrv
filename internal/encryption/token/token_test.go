package token

import (
	"strings"
	"testing"
	"time"
)

func TestTokenCreation(t *testing.T) {
	bearer := NewMD5DigestBearer("secure", 10*time.Second)
	issuer := "token-test"
	userId := uint64(1)
	email := "test@test.com"
	token := bearer.CreateToken(WithIssuer(issuer), WithSubject(userId), WithEmail(email))
	t.Logf("Token %s", token)
	if token == "" {
		t.Errorf("Invalid empty token has been created")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 &&
		(parts[0] == "" || parts[1] == "") {
		t.Errorf("Unexpected token format")
	}
	p, ok := bearer.GetPayload(token)
	if !ok {
		t.Errorf("Unexpected payload validation result (%v): %+v", ok, p)
	}
	if p.Issuer != issuer {
		t.Errorf("Unexpected issuer (%s): %s", issuer, p.Issuer)
	}
	if p.UserID != userId {
		t.Errorf("Unexpected user ID (%d): %d", userId, p.UserID)
	}
	if p.Email != email {
		t.Errorf("Unexpected email (%s): %s", email, p.Email)
	}

	// token without subject
}

func TestTokenExpiration(t *testing.T) {
	bearer := NewMD5DigestBearer("", 2*time.Second)
	token1 := bearer.CreateToken(WithSubject(1))
	t.Logf("Token #1 %s", token1)
	time.Sleep(1 * time.Second)
	if _, ok := bearer.GetPayload(token1); !ok {
		t.Errorf("Token unexpectedly expired before default expiration time")
	}
	time.Sleep(2 * time.Second)
	if _, ok := bearer.GetPayload(token1); ok {
		t.Errorf("Token is not expired after default expiration time")
	}
	// change default expiration
	token2 := bearer.CreateToken(
		WithSubject(1),
		WithExpiration(time.Now().UTC().Add(4*time.Second).Unix()),
	)
	t.Logf("Token #2 %s", token2)
	time.Sleep(3 * time.Second)
	if _, ok := bearer.GetPayload(token2); !ok {
		t.Errorf("Token unexpectedly expired before custom expiration time")
	}
	time.Sleep(2 * time.Second)
	if _, ok := bearer.GetPayload(token2); ok {
		t.Errorf("Token is not expired after custom expiration time")
	}

	if token1 == token2 {
		t.Errorf("Unexpected equal tokens.")
	}
}

func TestBearerEncryption(t *testing.T) {
	secure := NewMD5DigestBearer("secure", 1*time.Second)
	insecure := NewMD5DigestBearer("", 1*time.Second)
	expire := time.Now().UTC().Add(1 * time.Hour).Unix()
	secToken := secure.CreateToken(WithSubject(1), WithExpiration(expire))
	insToken := insecure.CreateToken(WithSubject(1), WithExpiration(expire))
	t.Logf("Secure token %s", secToken)
	t.Logf("Insecure token %s", insToken)
	secParts := strings.Split(secToken, ".")
	insParts := strings.Split(insToken, ".")
	if len(secParts) != len(insParts) &&
		len(secParts) != 2 {
		t.Errorf("Tokens have different format")
	}
	if secParts[0] != insParts[0] {
		t.Errorf("Tokens payload parts are not equal for the same payload")
	}
	if secParts[1] == insParts[1] {
		t.Errorf("Tokens of secure and insecure bearers have equal signature parts")
	}
	if _, ok := secure.GetPayload(secParts[0] + "." + insParts[1]); ok {
		t.Errorf("Bearer with secret does not use signature part of the token")
	}
	if _, ok := insecure.GetPayload(insParts[0] + "." + secParts[1]); ok {
		t.Errorf("Bearer without secret does not use signature part of the token")
	}
}
