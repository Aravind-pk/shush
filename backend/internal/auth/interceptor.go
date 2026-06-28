// Package auth verifies Clerk session tokens on incoming Connect RPCs.
//
// Clerk is the identity provider: the dashboard (and CLI) obtain a Clerk
// session JWT and send it as `Authorization: Bearer <token>`. This interceptor
// verifies that JWT against Clerk's JWKS and stashes the Clerk user id in the
// request context for handlers to use. We never issue our own tokens.
package auth

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
)

// ctxKey is unexported so only this package can set the value — handlers read
// it via UserID below.
type ctxKey struct{}

// NewInterceptor returns a Connect interceptor that requires a valid Clerk
// session token on every RPC. clerk.SetKey(...) must have been called first so
// the default JWKS client can fetch Clerk's signing keys.
func NewInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			authz := req.Header().Get("Authorization")
			token, ok := strings.CutPrefix(authz, "Bearer ")
			if !ok || token == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated,
					errors.New("missing Bearer token"))
			}

			claims, err := jwt.Verify(ctx, &jwt.VerifyParams{Token: token})
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated,
					errors.New("invalid session token"))
			}

			// claims.Subject is the Clerk user id, e.g. "user_2abc...".
			ctx = context.WithValue(ctx, ctxKey{}, claims.Subject)
			return next(ctx, req)
		}
	}
}

// UserID returns the authenticated Clerk user id placed in the context by the
// interceptor. ok is false if the request was not authenticated.
func UserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKey{}).(string)
	return id, ok && id != ""
}
