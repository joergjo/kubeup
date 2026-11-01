package main

import (
	"strings"
	"testing"

	"github.com/joergjo/kubeup/internal/event"
)

func TestSMTPConfiguration(t *testing.T) {
	complete := config{
		EmailFrom: "from@example.com", EmailTo: "to@example.com", EmailSubject: "AKS",
		SMTPHost: "smtp.example.com", SMTPPort: 587, SMTPUsername: "user", SMTPPassword: "password",
	}
	for _, tc := range []struct {
		name   string
		change func(*config)
		want   string
	}{
		{"complete", func(c *config) {}, ""},
		{"missing_host", func(c *config) { c.SMTPHost = "" }, "SMTP host required"},
		{"missing_username", func(c *config) { c.SMTPUsername = "" }, "SMTP username required"},
		{"missing_password", func(c *config) { c.SMTPPassword = "" }, "SMTP password required"},
		{"zero_port", func(c *config) { c.SMTPPort = 0 }, "SMTP port must be between"},
		{"negative_port", func(c *config) { c.SMTPPort = -1 }, "SMTP port must be between"},
		{"out_of_range_port", func(c *config) { c.SMTPPort = 65536 }, "SMTP port must be between"},
		{"missing_from", func(c *config) { c.EmailFrom = "" }, "email options required with SMTP"},
		{"missing_to", func(c *config) { c.EmailTo = "" }, "email options required with SMTP"},
		{"missing_subject", func(c *config) { c.EmailSubject = "" }, "email options required with SMTP"},
		{"host_only", func(c *config) { c.SMTPUsername, c.SMTPPassword = "", "" }, "SMTP username required"},
		{"username_only", func(c *config) { c.SMTPHost, c.SMTPPassword = "", "" }, "SMTP host required"},
		{"password_only", func(c *config) { c.SMTPHost, c.SMTPUsername = "", "" }, "SMTP host required"},
		{"logging_only", func(c *config) { *c = config{} }, ""},
		{"default_port_only", func(c *config) { *c = config{SMTPPort: 587} }, ""},
		{"custom_port_only", func(c *config) { *c = config{SMTPPort: 2525} }, ""},
		{"sendgrid_only", func(c *config) {
			c.SMTPHost, c.SMTPUsername, c.SMTPPassword = "", "", ""
			c.SendGridAPIKey = "test-key"
		}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := complete
			tc.change(&cfg)
			_, err := event.NewPublisher(publisherOptions(cfg)...)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("unexpected configuration error: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestRunRejectsIncompleteSMTP(t *testing.T) {
	if code := run(config{SMTPHost: "smtp.example.com", SMTPPort: 587}); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}
