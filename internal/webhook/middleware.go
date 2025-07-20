package webhook

import (
	"context"
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

func newClientSecretMiddleware(sec1, sec2 string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secret := r.URL.Query().Get(accessTokenParam)
			if secret != sec1 && secret != sec2 {
				slog.Warn("received request with invalid secret")
				http.Error(w, "invalid secret", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type roleClaims struct {
	Roles []string `json:"roles"`
}

func (c roleClaims) Validate(ctx context.Context) error {
	ok := slices.Contains(c.Roles, "AzureEventGridSecureWebhookSubscriber")
	if !ok {
		return fmt.Errorf("missing required role")
	}
	return nil
}

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
