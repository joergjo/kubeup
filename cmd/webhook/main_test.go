package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/joergjo/kubeup/internal/event"
	"github.com/joergjo/kubeup/internal/webhook"
)

func isolateConfiguration(t *testing.T, args ...string) {
	t.Helper()
	for _, name := range []string{
		"KU_PORT", "KU_PATH", "KU_DEBUG", "KU_EMAIL_FROM", "KU_EMAIL_TO", "KU_EMAIL_SUBJECT",
		"KU_SMTP_HOST", "KU_SMTP_PORT", "KU_SMTP_USERNAME", "KU_SMTP_PASSWORD", "KU_SENDGRID_APIKEY",
		"KU_TENANT_ID", "KU_APP_ID", "KU_SECRET_1", "KU_SECRET_2",
	} {
		value, present := os.LookupEnv(name)
		t.Cleanup(func() {
			if present {
				if err := os.Setenv(name, value); err != nil {
					t.Error(err)
				}
			} else if err := os.Unsetenv(name); err != nil {
				t.Error(err)
			}
		})
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
	}
	oldFlags, oldArgs := flag.CommandLine, os.Args
	t.Cleanup(func() { flag.CommandLine, os.Args = oldFlags, oldArgs })
	flag.CommandLine = flag.NewFlagSet("configuration-test", flag.ContinueOnError)
	os.Args = append([]string{"configuration-test"}, args...)
}

func TestConfigure(t *testing.T) {
	defaults := config{Port: 8000, Path: "/webhook", SMTPPort: 587}
	for _, tc := range []struct {
		name string
		args []string
		env  map[string]string
		want config
	}{
		{"defaults", nil, nil, defaults},
		{"flags", []string{"-port", "8080", "-path", "/events", "-debug"}, nil,
			config{Port: 8080, Path: "/events", Debug: true, SMTPPort: 587}},
		{"explicit_defaults", nil, map[string]string{"KU_PORT": "8000", "KU_PATH": "/webhook", "KU_DEBUG": "false"}, defaults},
		{"environment_precedence", []string{"-port", "8080", "-path", "/ignored"},
			map[string]string{"KU_PORT": "9000", "KU_PATH": "/events", "KU_DEBUG": "true"},
			config{Port: 9000, Path: "/events", Debug: true, SMTPPort: 587}},
		{"all_environment_fields", nil, map[string]string{
			"KU_PORT": "9001", "KU_PATH": "/custom", "KU_DEBUG": "true",
			"KU_EMAIL_FROM": "from@example.com", "KU_EMAIL_TO": "to@example.com", "KU_EMAIL_SUBJECT": "AKS",
			"KU_SMTP_HOST": "smtp.example.com", "KU_SMTP_PORT": "2525", "KU_SMTP_USERNAME": "user", "KU_SMTP_PASSWORD": "test-password",
			"KU_SENDGRID_APIKEY": "test-key", "KU_TENANT_ID": "tenant", "KU_APP_ID": "app",
			"KU_SECRET_1": "first", "KU_SECRET_2": "second",
		}, config{Port: 9001, Path: "/custom", Debug: true, EmailFrom: "from@example.com", EmailTo: "to@example.com",
			EmailSubject: "AKS", SMTPHost: "smtp.example.com", SMTPPort: 2525, SMTPUsername: "user", SMTPPassword: "test-password",
			SendGridAPIKey: "test-key", TenantID: "tenant", AppID: "app", Secret1: "first", Secret2: "second"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			isolateConfiguration(t, tc.args...)
			for name, value := range tc.env {
				t.Setenv(name, value)
			}
			if got := configure(); got != tc.want {
				t.Errorf("configuration = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestConfigureInvalidEnvironment(t *testing.T) {
	for _, tc := range []struct{ name, field string }{
		{"KU_PORT", "Port"}, {"KU_SMTP_PORT", "SMTPPort"}, {"KU_DEBUG", "Debug"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			isolateConfiguration(t)
			t.Setenv(tc.name, "invalid")
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(binary, "-test.run=^TestConfigureProcess$")
			cmd.Env = append(os.Environ(), "KUBEUP_CONFIG_TEST_HELPER=1")
			output, err := cmd.CombinedOutput()
			exitErr, ok := err.(*exec.ExitError)
			if !ok || exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit 1, got %v; output: %s", err, output)
			}
			if !strings.Contains(string(output), "parsing environment variables") || !strings.Contains(string(output), tc.field) {
				t.Fatalf("missing configuration diagnostic: %s", output)
			}
		})
	}
}

func TestConfigureProcess(t *testing.T) {
	if os.Getenv("KUBEUP_CONFIG_TEST_HELPER") != "1" {
		return
	}
	flag.CommandLine = flag.NewFlagSet("configuration-helper", flag.ContinueOnError)
	os.Args = []string{"configuration-helper"}
	configure()
}

func TestConfigurationPredicates(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  config
		want [4]bool
	}{
		{"absent", config{}, [4]bool{}},
		{"complete", config{EmailFrom: "from", EmailTo: "to", EmailSubject: "subject", SendGridAPIKey: "key",
			TenantID: "tenant", AppID: "app", Secret1: "first", Secret2: "second"}, [4]bool{true, true, true, true}},
		{"missing_from", config{EmailTo: "to", EmailSubject: "subject"}, [4]bool{}},
		{"missing_to", config{EmailFrom: "from", EmailSubject: "subject"}, [4]bool{}},
		{"missing_subject", config{EmailFrom: "from", EmailTo: "to"}, [4]bool{}},
		{"tenant_only", config{TenantID: "tenant"}, [4]bool{}},
		{"app_only", config{AppID: "app"}, [4]bool{}},
		{"first_secret_only", config{Secret1: "first"}, [4]bool{}},
		{"second_secret_only", config{Secret2: "second"}, [4]bool{}},
		{"sendgrid_only", config{SendGridAPIKey: "key"}, [4]bool{false, true, false, false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := [4]bool{tc.cfg.HasEmail(), tc.cfg.HasSendGrid(), tc.cfg.HasEntraID(), tc.cfg.HasClientSecrets()}
			if got != tc.want {
				t.Errorf("email/sendgrid/entra/secrets = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNotificationConfiguration(t *testing.T) {
	email := config{EmailFrom: "from@example.com", EmailTo: "to@example.com", EmailSubject: "AKS"}
	for _, tc := range []struct {
		name   string
		change func(*config)
		count  int
		want   string
	}{
		{"email_without_transport", func(c *config) {}, 2, ""},
		{"logging_only", func(c *config) { *c = config{} }, 1, ""},
		{"sendgrid", func(c *config) { c.SendGridAPIKey = "test-key" }, 3, ""},
		{"sendgrid_missing_from", func(c *config) { c.SendGridAPIKey, c.EmailFrom = "test-key", "" }, 2, "email options required with SendGrid"},
		{"sendgrid_missing_to", func(c *config) { c.SendGridAPIKey, c.EmailTo = "test-key", "" }, 2, "email options required with SendGrid"},
		{"sendgrid_missing_subject", func(c *config) { c.SendGridAPIKey, c.EmailSubject = "test-key", "" }, 2, "email options required with SendGrid"},
		{"both_transports", func(c *config) {
			c.SendGridAPIKey = "test-key"
			c.SMTPHost, c.SMTPPort, c.SMTPUsername, c.SMTPPassword = "smtp.example.com", 587, "user", "test-password"
		}, 4, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := email
			tc.change(&cfg)
			opts := publisherOptions(cfg)
			if len(opts) != tc.count {
				t.Fatalf("option count = %d, want %d", len(opts), tc.count)
			}
			_, err := event.NewPublisher(opts...)
			if tc.want == "" && err != nil || tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
				t.Errorf("expected error %q, got %v", tc.want, err)
			}
		})
	}
}

func TestServerConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  config
		want int
	}{
		{"unprotected", config{}, 0},
		{"secrets", config{Secret1: "first", Secret2: "second"}, 1},
		{"entra", config{TenantID: "tenant", AppID: "app"}, 1},
		{"entra_precedence", config{TenantID: "tenant", AppID: "app", Secret1: "first", Secret2: "second"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := serverOptions(tc.cfg)
			if len(opts) != tc.want {
				t.Fatalf("option count = %d, want %d", len(opts), tc.want)
			}
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
			server, err := webhook.NewServer(handler, append(opts, webhook.WithPath("/events"), webhook.WithPort(8088))...)
			if err != nil {
				t.Fatal(err)
			}
			if server.Addr != ":8088" {
				t.Errorf("address = %s, want :8088", server.Addr)
			}
			for _, token := range []string{"", "invalid", "first", "second"} {
				req := httptest.NewRequest(http.MethodPost, "/events?access_token="+token, nil)
				res := httptest.NewRecorder()
				server.Handler.ServeHTTP(res, req)
				want := http.StatusNoContent
				if tc.cfg.HasEntraID() || tc.cfg.HasClientSecrets() && token != "first" && token != "second" {
					want = http.StatusUnauthorized
				}
				if res.Code != want {
					t.Errorf("token %q status = %d, want %d", token, res.Code, want)
				}
			}
			res := httptest.NewRecorder()
			server.Handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/webhook", nil))
			if res.Code != http.StatusNotFound {
				t.Errorf("default route status = %d, want 404", res.Code)
			}
		})
	}
}

func TestHTTPConfigurationErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
		port int
		want string
	}{
		{"empty_path", "", 8000, "path required"},
		{"zero_port", "/events", 0, "port required"},
		{"negative_port", "/events", -1, "port required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := webhook.NewServer(http.NotFoundHandler(), webhook.WithPath(tc.path), webhook.WithPort(tc.port))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("expected %q, got %v", tc.want, err)
			}
		})
	}
}

func TestDebugLoggingConfiguration(t *testing.T) {
	for _, debug := range []bool{false, true} {
		t.Run(map[bool]string{false: "production", true: "debug"}[debug], func(t *testing.T) {
			original := slog.Default()
			t.Cleanup(func() { slog.SetDefault(original) })
			flush := setDefaultLogger(debug)
			t.Cleanup(flush)
			if got := slog.Default().Enabled(context.Background(), slog.LevelDebug); got != debug {
				t.Errorf("debug logging enabled = %v, want %v", got, debug)
			}
			if !slog.Default().Enabled(context.Background(), slog.LevelInfo) {
				t.Error("info logging must remain enabled")
			}
		})
	}
}

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
	original := slog.Default()
	t.Cleanup(func() { slog.SetDefault(original) })
	if code := run(config{SMTPHost: "smtp.example.com", SMTPPort: 587}); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}
