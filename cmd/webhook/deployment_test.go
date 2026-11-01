package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestAzureDeploymentStatus(t *testing.T) {
	mock := `#!/bin/bash
case "$*" in
  *"--template-file webhook.bicep"*) echo "webhook.example.invalid" ;;
  *"--template-file eventgrid.bicep"*) exit "$MOCK_EVENTGRID_STATUS" ;;
esac
`
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "az"), []byte(mock), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		cluster string
		status  int
		message string
	}{
		{"success", "cluster", 0, "Event Grid topic has been created successfully."},
		{"failure", "cluster", 42, "Failed to deploy Event Grid subscription."},
		{"skipped", "", 0, "Skipping Event Grid topic creation."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("bash", "../../deploy/azure.sh")
			cmd.Env = []string{
				"PATH=" + dir + ":" + os.Getenv("PATH"),
				"KU_RESOURCE_GROUP=webhook-rg",
				"KU_AKS_RESOURCE_GROUP=aks-rg",
				"KU_AKS_CLUSTER=" + tc.cluster,
				"MOCK_EVENTGRID_STATUS=" + strconv.Itoa(tc.status),
			}
			output, err := cmd.CombinedOutput()
			status := 0
			if err != nil {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatal(err)
				}
				status = exitErr.ExitCode()
			}
			if status != tc.status {
				t.Errorf("exit status = %d, want %d; output: %s", status, tc.status, output)
			}
			if !strings.Contains(string(output), tc.message) {
				t.Errorf("missing %q in output: %s", tc.message, output)
			}
			if tc.name != "success" && strings.Contains(string(output), "Event Grid topic has been created successfully.") {
				t.Errorf("unexpected success message: %s", output)
			}
		})
	}
}
