package main

import (
	"log/slog"
	"strings"
	"testing"
)

func TestRejectInvalidAuthorizationConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  config
		want string
	}{
		{"absent", config{}, "authorization required"},
		{"tenant_only", config{TenantID: "tenant"}, "KU_APP_ID"},
		{"app_only", config{AppID: "app"}, "KU_TENANT_ID"},
		{"first_secret_only", config{Secret1: "first"}, "KU_SECRET_2"},
		{"second_secret_only", config{Secret2: "second"}, "KU_SECRET_1"},
		{"partial_entra_with_secrets", config{TenantID: "tenant", Secret1: "first", Secret2: "second"}, "KU_APP_ID"},
		{"partial_secrets_with_entra", config{TenantID: "tenant", AppID: "app", Secret1: "first"}, "KU_SECRET_2"},
		{"partial_entra_insecure", config{TenantID: "tenant", Insecure: true}, "KU_APP_ID"},
		{"partial_secrets_insecure", config{Secret1: "first", Insecure: true}, "KU_SECRET_2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := serverOptions(tc.cfg)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, err)
			}
			original := slog.Default()
			t.Cleanup(func() { slog.SetDefault(original) })
			if code := run(tc.cfg); code != 1 {
				t.Fatalf("startup exit code = %d, want 1", code)
			}
		})
	}
}
