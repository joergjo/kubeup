package main

import (
	"os"
	"reflect"
	"regexp"
	"testing"
)

func TestDeploymentEnvironmentVariables(t *testing.T) {
	bicep, err := os.ReadFile("../../deploy/modules/app.bicep")
	if err != nil {
		t.Fatal(err)
	}

	configType := reflect.TypeFor[config]()
	envNames := make(map[string]bool)
	for i := range configType.NumField() {
		envNames[configType.Field(i).Tag.Get("env")] = true
	}

	pattern := regexp.MustCompile(`(?m)^\s*name:\s*'(KU_[A-Z0-9_]+)'`)
	matches := pattern.FindAllSubmatch(bicep, -1)
	if len(matches) == 0 {
		t.Fatal("no application environment variables found in deployment template")
	}
	for _, match := range matches {
		name := string(match[1])
		if !envNames[name] {
			t.Errorf("deployment environment variable %q is not recognized by application configuration", name)
		}
	}
}
