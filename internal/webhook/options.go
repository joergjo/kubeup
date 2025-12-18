package webhook

import (
	"errors"
	"fmt"
	"net/url"
)

type options struct {
	path         string
	port         int
	clientSecret *clientSecretOptions
	entraID      *entraIDOptions
}

type entraIDOptions struct {
	issuerURL *url.URL
	appID     string
}

type clientSecretOptions struct {
	secret1 string
	secret2 string
}

// Options represents a functional option for the webhook.
type Options func(options *options) error

// WithClientSecret configures the webhook to use a specific client secret.
func WithClientSecret(sec1, sec2 string) Options {
	return func(options *options) error {
		if sec1 == "" || sec2 == "" {
			return errors.New("two client secrets required")
		}
		options.clientSecret = &clientSecretOptions{secret1: sec1, secret2: sec2}
		return nil
	}
}

// WithEntraID configures the webhook to check for a valid access token issued by Entra ID.
func WithEntraID(tenantID, appID string) Options {
	return func(options *options) error {
		iss := fmt.Sprintf("https://sts.windows.net/%s/", tenantID)
		issURL, err := url.Parse(iss)
		if err != nil {
			return fmt.Errorf("issuer URL required: %w", err)
		}
		if appID == "" {
			return errors.New("app ID URI required")
		}
		options.entraID = &entraIDOptions{issuerURL: issURL, appID: appID}
		return nil
	}
}

// WithPath configures the webhook to use a specific path.
func WithPath(path string) Options {
	return func(options *options) error {
		if path == "" {
			return errors.New("path required")
		}
		options.path = path
		return nil
	}
}

// WithPort configures the webhook to use a specific port.
func WithPort(port int) Options {
	return func(options *options) error {
		if port <= 0 {
			return errors.New("port required")
		}
		options.port = port
		return nil
	}
}
