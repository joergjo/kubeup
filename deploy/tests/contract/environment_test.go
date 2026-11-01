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
