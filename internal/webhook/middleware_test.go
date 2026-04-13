package webhook

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/joergjo/kubeup/internal/event"
)

// Test tokens generated with http://jwtbuilder.jamiekurtz.com
const (
	validToken            = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vand0YnVpbGRlci5qYW1pZWt1cnR6LmNvbSIsImlhdCI6MTcyNjIxMDE3NSwiZXhwIjoyMDQxNzQyOTc1LCJhdWQiOiJnaXRodWIuY29tL2pvZXJnam8va3ViZXVwIiwic3ViIjoidGVzdCIsInJvbGVzIjpbIkF6dXJlRXZlbnRHcmlkU2VjdXJlV2ViaG9va1N1YnNjcmliZXIiLCJUZXN0Il19.6t1hMndbSUkfxU2Ly9af20O5RCBZZQMva1akE4e3ktU"
	validTokenMissingRole = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vand0YnVpbGRlci5qYW1pZWt1cnR6LmNvbSIsImlhdCI6MTcyNjIxMDE3NSwiZXhwIjoyMDQxNzQyOTc1LCJhdWQiOiJnaXRodWIuY29tL2pvZXJnam8va3ViZXVwIiwic3ViIjoidGVzdCIsInJvbGVzIjpbIlRlc3QxIiwiVGVzdDIiXX0.QHqha2rbhP41qSHO986q2B0v7qgrXrqWRQExHt7Zdjs"
	invalidToken          = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vand0YnVpbGRlci5qYW1pZWt1cnR6LmNvbSIsImlhdCI6MTcyNjIxMDE3NSwiZXhwIjoyMDQxNzQyOTc1LCJhdWQiOiJnaXRodWIuY29tL2pvZXJnam8va3ViZXVwL2ludmFsaWQiLCJzdWIiOiJ0ZXN0Iiwicm9sZXMiOlsiQXp1cmVFdmVudEdyaWRTZWN1cmVXZWJob29rU3Vic2NyaWJlciIsIlRlc3QiXX0.4wSl2QYyZVvhf1Fv52AaEVT2DM5roKCvuHFS7PVuIrI"
	signingKey            = "qwertyuiopasdfghjklzxcvbnm123456"
	issuer                = "http://jwtbuilder.jamiekurtz.com"
	audience              = "github.com/joergjo/kubeup"
	origin                = "eventgrid.azure.net"
)

func TestClientSecretMiddleware(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		status int
	}{
		{
			name:   "valid_first_secret",
			secret: "secret1",
			status: http.StatusOK,
		},
		{
			name:   "valid_second_secret",
			secret: "secret2",
			status: http.StatusOK,
		},
		{
			name:   "invalid_secret",
			secret: "invalid",
			status: http.StatusUnauthorized,
		},
		{
			name:   "missing_secret",
			secret: "",
			status: http.StatusUnauthorized,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := event.NewPublisher()
			h, err := NewCloudEventHandler(context.Background(), p)
			if err != nil {
				t.Fatalf("Error creating handler: %v", err)
			}
			url := fmt.Sprintf("/webhook?%s=%s", accessTokenParam, tc.secret)
			req := httptest.NewRequest(http.MethodOptions, url, nil)
			req.Header.Set("WebHook-Request-Origin", origin)
			res := httptest.NewRecorder()

			mw := newClientSecretMiddleware("secret1", "secret2")(h)
			mw.ServeHTTP(res, req)
			if res.Result().StatusCode != tc.status {
				t.Errorf("Want status code %d, got %d", tc.status, res.Result().StatusCode)
			}
		})
	}
}

func TestAccessTokenMiddleware(t *testing.T) {
	tests := []struct {
		name   string
		token  string
		status int
	}{
		{
			name:   "valid_token",
			token:  validToken,
			status: http.StatusOK,
		},
		{
			name:   "valid_token_missing_role",
			token:  validTokenMissingRole,
			status: http.StatusUnauthorized,
		},
		{
			name:   "invalid_token",
			token:  invalidToken,
			status: http.StatusUnauthorized,
		},
		{
			name:   "missing_token",
			token:  "",
			status: http.StatusUnauthorized,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := event.NewPublisher()
			h, err := NewCloudEventHandler(context.Background(), p)
			if err != nil {
				t.Fatalf("Error creating handler: %v", err)
			}
			v, _ := newTestValidator(issuer, audience, signingKey)
			mw, err := newEntraIDMiddleware(v)
			if err != nil {
				t.Errorf("Error creating middleware: %v", err)
			}
			req := httptest.NewRequest(http.MethodOptions, "/webhook", nil)
			req.Header.Set("WebHook-Request-Origin", origin)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.token))
			res := httptest.NewRecorder()
			mw(h).ServeHTTP(res, req)
			if res.Result().StatusCode != tc.status {
				t.Errorf("Want status code %d, got %d", tc.status, res.Result().StatusCode)
			}
		})
	}
}

func newTestValidator(iss, aud, key string) (*validator.Validator, error) {
	v, err := validator.New(validator.WithAlgorithm(validator.HS256),
		validator.WithIssuer(iss),
		validator.WithAudience(aud),
		validator.WithKeyFunc(func(ctx context.Context) (any, error) {
			return []byte(key), nil
		}),
		validator.WithCustomClaims(func() validator.CustomClaims {
			return &roleClaims{}
		}),
		validator.WithAllowedClockSkew(time.Minute))
	if err != nil {
		return nil, err
	}
	return v, nil
}
