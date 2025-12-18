package webhook

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// Defined in https://github.com/cloudevents/spec/blob/main/cloudevents/http-webhook.md#32-uri-query-parameter.
const accessTokenParam = "access_token"

// newClientSecretMiddleware creates a middleware that validates the client secret in request query parameters.
// It expects the secret in the "access_token" query parameter and validates it against two possible values (for rotation).
func newClientSecretMiddleware(key1, key2 string) func(http.Handler) http.Handler {
	key1B := []byte(key1)
	key2B := []byte(key2)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenB := []byte(r.URL.Query().Get(accessTokenParam))
			if subtle.ConstantTimeCompare(key1B, tokenB) != 1 && subtle.ConstantTimeCompare(key2B, tokenB) != 1 {
				slog.Warn("received request with invalid secret")
				http.Error(w, "invalid secret", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// roleClaims represents the custom claims containing roles from an Entra ID JWT.
type roleClaims struct {
	Roles []string `json:"roles"`
}

// Validate checks if the required role "AzureEventGridSecureWebhookSubscriber" is present in the claims.
// This implements the validator.CustomClaims interface.
func (c roleClaims) Validate(ctx context.Context) error {
	ok := slices.Contains(c.Roles, "AzureEventGridSecureWebhookSubscriber")
	if !ok {
		return fmt.Errorf("missing required role")
	}
	return nil
}

// newEntraIDMiddleware creates a middleware that validates Entra ID JWT tokens.
// It uses the provided validator to check token authenticity and required claims.
func newEntraIDMiddleware(v *validator.Validator) func(http.Handler) http.Handler {
	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("validating token", "error", err)
		w.WriteHeader(http.StatusUnauthorized)
	}

	middleware := jwtmiddleware.New(
		v.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return middleware.CheckJWT(next)
	}
}

// newValidator creates a new JWT validator for Entra ID tokens.
// It validates token signature, issuer, audience, and custom role claims.
func newValidator(iss *url.URL, aud string) (*validator.Validator, error) {
	p := jwks.NewCachingProvider(iss, 5*time.Minute)
	v, err := validator.New(
		p.KeyFunc,
		validator.RS256,
		iss.String(),
		[]string{aud},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &roleClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return nil, err
	}
	return v, nil
}
