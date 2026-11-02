//go:build deployment_contract

package contract

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestDeploymentEnvironmentVariables(t *testing.T) {
	source, err := parser.ParseFile(token.NewFileSet(), "../../../cmd/webhook/main.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	envNames := make(map[string]bool)
	ast.Inspect(source, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if !ok || spec.Name.Name != "config" {
			return true
		}
		structure, ok := spec.Type.(*ast.StructType)
		if !ok {
			t.Fatal("config is not a struct")
		}
		for _, field := range structure.Fields.List {
			if field.Tag == nil {
				continue
			}
			tag, err := strconv.Unquote(field.Tag.Value)
			if err != nil {
				t.Fatal(err)
			}
			if name := reflect.StructTag(tag).Get("env"); name != "" {
				envNames[name] = true
			}
		}
		return false
	})
	if len(envNames) == 0 {
		t.Fatal("no application configuration environment tags found")
	}
	bicep, err := os.ReadFile("../../modules/app.bicep")
	if err != nil {
		t.Fatal(err)
	}
	matches := regexp.MustCompile(`(?m)^\s*name:\s*'(KU_[A-Z0-9_]+)'`).FindAllSubmatch(bicep, -1)
	if len(matches) == 0 {
		t.Fatal("no application environment variables found in deployment template")
	}
	for _, match := range matches {
		if name := string(match[1]); !envNames[name] {
			t.Errorf("deployment variable %q is not recognized by application configuration", name)
		}
	}
}

func TestAzureMonitorLoggingConfiguration(t *testing.T) {
	environment, err := os.ReadFile("../../modules/environment.bicep")
	if err != nil {
		t.Fatal(err)
	}
	app, err := os.ReadFile("../../modules/app.bicep")
	if err != nil {
		t.Fatal(err)
	}

	for _, check := range []struct {
		name string
		text string
		want string
	}{
		{
			name: "managed environment uses the GA API version",
			text: string(environment),
			want: "Microsoft.App/managedEnvironments@2026-07-01",
		},
		{
			name: "container app uses the GA API version",
			text: string(app),
			want: "Microsoft.App/containerApps@2026-07-01",
		},
		{
			name: "environment uses Azure Monitor logging",
			text: string(environment),
			want: "destination: 'azure-monitor'",
		},
		{
			name: "diagnostic setting targets the environment",
			text: string(environment),
			want: "scope: environment",
		},
		{
			name: "diagnostic setting targets the Log Analytics workspace",
			text: string(environment),
			want: "workspaceId: logAnalyticsWorkspace.id",
		},
		{
			name: "diagnostic setting uses the newest supported API version",
			text: string(environment),
			want: "Microsoft.Insights/diagnosticSettings@2021-05-01-preview",
		},
	} {
		t.Run(check.name, func(t *testing.T) {
			if !strings.Contains(check.text, check.want) {
				t.Errorf("Bicep template does not contain %q", check.want)
			}
		})
	}

	for _, category := range []string{
		"ContainerAppSystemLogs",
		"ContainerAppConsoleLogs",
		"ContainerAppHTTPLogs",
	} {
		if !regexp.MustCompile(`(?s)category:\s*'` + regexp.QuoteMeta(category) + `'\s+enabled:\s*true`).Match(environment) {
			t.Errorf("environment diagnostic setting does not enable %s", category)
		}
	}
	if strings.Contains(string(environment), "dynamicJsonColumns") {
		t.Error("environment Bicep still configures dynamicJsonColumns")
	}
}
