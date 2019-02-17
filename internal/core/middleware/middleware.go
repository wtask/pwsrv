package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/wtask/pwsrv/internal/core/reply"
)

type contextKey int

const (
	_ contextKey = iota
	userIDKey
)

type (
	TokenDiscoverer interface {
		DiscoverUserID(token string) (uint64, bool)
	}
)

// TODO Add SupplyRecovery middleware to handle panics inside end-points.

// AuthorizationTryout - generates middleware which attempts to supply user from Authorization header.
func AuthorizationTryout(b TokenDiscoverer) func(http.Handler) http.Handler {
	if b == nil {
		panic(errors.New("middleware.AuthorizationTryout: TokenDiscoverer is nil"))
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if b != nil {
				token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
				if userID, ok := b.DiscoverUserID(token); ok && userID > 0 {
					ctx := context.WithValue(r.Context(), userIDKey, userID)
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
			if userID, ok := DiscoverUserID(r); !ok || userID == 0 {
				reply.Unauthorized()(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// DiscoverUserID - return user ID from request and existence flag.
func DiscoverUserID(r *http.Request) (uint64, bool) {
	v, ok := r.Context().Value(userIDKey).(uint64)
	return v, ok
}
