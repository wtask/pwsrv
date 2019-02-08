package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/wtask/pwsrv/internal/encryption/token"

	"github.com/wtask/pwsrv/internal/core/reply"
)

type contextKey int

const (
	_ contextKey = iota
	authPayload
)

// TODO Add SupplyRecovery middleware to handle panics inside end-points.

// AuthorizationTryout - generates middleware which attempts to supply user from Authorization header.
func AuthorizationTryout(b token.AuthBearer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if b != nil {
				token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
				if p, ok := b.GetPayload(token); ok {
					ctx := context.WithValue(r.Context(), authPayload, p)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AuthorizationRequired - generates middleware to check User exist within request context.
// If it not exist, sets Unauthorized status for response and terminates middleware chain.
func AuthorizationRequired() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if p := GetAuthPayload(r); p == nil || p.UserID == 0 {
				reply.Unauthorized()(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetAuthPayload - get authorization payload data from request context.
func GetAuthPayload(r *http.Request) *token.Payload {
	p, ok := r.Context().Value(authPayload).(*token.Payload)
	if !ok || p == nil {
		return nil
	}
	return p
}
