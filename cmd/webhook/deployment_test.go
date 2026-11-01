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

func TestEntraSetupReusesSavedConfiguration(t *testing.T) {
	script, err := os.ReadFile("../../deploy/entraid.sh")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	deploy := filepath.Join(root, "deploy")
	if err := os.Mkdir(deploy, 0o700); err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(deploy, "entraid.sh")
	if err := os.WriteFile(scriptPath, script, 0o700); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, ".env.entraid")
	const saved = "KU_APP_ID='existing-application'\n"
	if err := os.WriteFile(configPath, []byte(saved), 0o600); err != nil {
		t.Fatal(err)
	}
	calls := filepath.Join(root, "graph-calls")
	mock := "#!/bin/bash\nprintf 'unexpected Graph call\\n' >> \"$CALL_LOG\"\nexit 42\n"
	if err := os.WriteFile(filepath.Join(root, "mgc"), []byte(mock), 0o700); err != nil {
		t.Fatal(err)
	}
	for i := range 2 {
		cmd := exec.Command("bash", scriptPath)
		cmd.Dir = deploy
		cmd.Env = []string{"PATH=" + root + ":" + os.Getenv("PATH"), "CALL_LOG=" + calls}
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("invocation %d failed: %v; output: %s", i+1, err, output)
		}
		if string(output) != saved {
			t.Fatalf("invocation %d output = %q, want %q", i+1, output, saved)
		}
	}
	if _, err := os.Stat(calls); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no Graph calls, got stat error: %v", err)
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != saved {
		t.Fatalf("saved configuration was changed: %q", content)
	}
}

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
