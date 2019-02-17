package token

import (
	"strings"
	"testing"
	"time"
)

func TestMD5DigestTokenCreation(t *testing.T) {
	cases := []struct {
		options []bearerOption
		userID  uint64
	}{
		{
			[]bearerOption{WithSignatureSecret("secret"), WithTTL(1 * time.Minute), WithIssuer("token-test")}, 1,
		},
		{
			// invalid user ID, but valid uint64
			[]bearerOption{WithSignatureSecret("secret"), WithTTL(1 * time.Minute), WithIssuer("token-test")}, 0,
		},
		{
			// with empty secret and without issuer
			[]bearerOption{WithTTL(1 * time.Minute)}, 1,
		},
		{
			// synthetic test - nothing will pass to AuthBear factory
			[]bearerOption{}, 1,
		},
		{
			// all unusable options are ignored
			[]bearerOption{nil}, 1,
		},
		{
			// all unusable options are ignored
			[]bearerOption{nil, nil, nil}, 1,
		},
		{
			// with custom time-provider
			[]bearerOption{withTimeProvider(func() time.Time { return time.Time{} })}, 1,
		},
		{
			// given time-provider is ignored
			[]bearerOption{withTimeProvider(nil)}, 1,
		},
	}
	for _, c := range cases {
		bearer := NewMD5DigestBearer(c.options...)
		t.Logf("%s", time.Now().Format("2006-01-02 15:04:05.000000000"))
		token := bearer.NewToken(c.userID)

		t.Logf("\ttoken %s", token)
		if token == "" {
			t.Errorf("Empty token has been created")
		}

		parts := strings.Split(token, ".")
		if len(parts) != 2 &&
			(parts[0] == "" || parts[1] == "") {
			t.Errorf("Unexpected token format")
		}
	}
}

func TestTokenExpiration(t *testing.T) {
	now := time.Now()
	bearer := NewMD5DigestBearer(
		WithTTL(10*time.Second),
		withTimeProvider(func() time.Time { return now }),
	)
	token1 := bearer.NewToken(1)
	t.Logf("Token #1 %s", token1)
	now = now.Add(5 * time.Second)
	if _, ok := bearer.DiscoverUserID(token1); !ok {
		t.Errorf("Token unexpectedly expired while TTL goes on")
	}
	now = now.Add(5 * time.Second)
	if _, ok := bearer.DiscoverUserID(token1); ok {
		t.Errorf("Token is not expired as expected")
	}
	token2 := bearer.NewToken(1)
	t.Logf("Token #2 %s", token2)
	if token1 == token2 {
		t.Errorf("Unexpected equal tokens.")
	}
}

func TestBearerEncryption(t *testing.T) {
	now := time.Now()
	secure := NewMD5DigestBearer(
		WithSignatureSecret("secret"),
		WithTTL(5*time.Second),
		withTimeProvider(func() time.Time { return now }),
	)
	insecure := NewMD5DigestBearer(
		WithTTL(5*time.Second),
		withTimeProvider(func() time.Time { return now }),
	)
	secToken := secure.NewToken(1)
	insToken := insecure.NewToken(1)
	t.Logf("Secure token %s", secToken)
	t.Logf("Insecure token %s", insToken)
	if secToken == insToken {
		t.Errorf("Secure and insecure tokens are equal")
	}
	secParts := strings.Split(secToken, ".")
	insParts := strings.Split(insToken, ".")
	if len(secParts) != 2 ||
		len(insParts) != 2 {
		t.Errorf("Tokens have wrong format")
	}
	if secParts[0] != insParts[0] {
		t.Errorf("Payload parts are not equal for the same payload")
	}
	if secParts[1] == insParts[1] {
		t.Errorf("Signature parts of secure and insecure tokens are equal")
	}
	if _, ok := secure.DiscoverUserID(secParts[0] + "." + insParts[1]); ok {
		t.Errorf("Bearer with secret does not use signature part of the token")
	}
	if _, ok := insecure.DiscoverUserID(insParts[0] + "." + secParts[1]); ok {
		t.Errorf("Bearer without secret does not use signature part of the token")
	}
}
