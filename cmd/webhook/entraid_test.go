package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testAppID       = "11111111-1111-1111-1111-111111111111"
	testObjectID    = "22222222-2222-2222-2222-222222222222"
	testWebhookSP   = "33333333-3333-3333-3333-333333333333"
	testEventGridSP = "44444444-4444-4444-4444-444444444444"
	testUserID      = "55555555-5555-5555-5555-555555555555"
	testRoleID      = "abcdefab-abcd-abcd-abcd-abcdefabcdef"
)

func TestEntraGraphSetup(t *testing.T) {
	if _, err := exec.LookPath("jq"); err != nil {
		t.Fatal("deployment tests require jq")
	}
	script, err := os.ReadFile("../../deploy/entraid.sh")
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"new", "existing", "paginated", "propagation", "unknown-app", "no-login", "non-user", "forbidden", "missing-id", "malformed-lookup", "create-failure", "principal-missing-id", "patch-failure", "assignment-failure", "no-role"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			deploy := filepath.Join(root, "deploy")
			if err := os.Mkdir(deploy, 0o700); err != nil {
				t.Fatal(err)
			}
			for name, content := range map[string]string{
				"deploy/entraid.sh": string(script),
				"az":                "#!/bin/bash\nexec \"$HELPER_BINARY\" -test.run=TestGraphCLIHelper -- \"$@\"\n",
				"uuidgen":           "#!/bin/bash\necho " + strings.ToUpper(testRoleID) + "\n",
				"sleep":             "#!/bin/bash\nexit 0\n",
			} {
				if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{filepath.Join(deploy, "entraid.sh")}
			if mode == "existing" || mode == "paginated" || mode == "unknown-app" {
				args = append(args, testAppID)
			}
			cmd := exec.Command("bash", args...)
			cmd.Dir = deploy
			cmd.Env = append(os.Environ(), "PATH="+root+":"+os.Getenv("PATH"),
				"GRAPH_HELPER=1", "HELPER_BINARY="+binary, "GRAPH_MODE="+mode,
				"GRAPH_CALLS="+filepath.Join(root, "calls"))
			output, err := cmd.CombinedOutput()
			success := mode == "new" || mode == "existing" || mode == "paginated" || mode == "propagation"
			if (err == nil) != success {
				log, _ := os.ReadFile(filepath.Join(deploy, "entraid.log"))
				t.Fatalf("success=%v, err=%v, output=%s, log=%s", success, err, output, log)
			}
			config, configErr := os.ReadFile(filepath.Join(root, ".env.entraid"))
			if !success {
				if !os.IsNotExist(configErr) {
					t.Fatal("failed setup wrote configuration")
				}
				if len(output) == 0 {
					t.Fatal("failed setup did not report an error")
				}
				return
			}
			if configErr != nil || string(config) != "KU_APP_ID='"+testAppID+"'\n" {
				t.Fatalf("unexpected configuration: %q, %v", config, configErr)
			}
			calls, err := os.ReadFile(filepath.Join(root, "calls"))
			if err != nil {
				t.Fatal(err)
			}
			if mode == "new" || mode == "propagation" {
				for _, expected := range []string{"/applications", testObjectID, `"principalId":"` + testUserID, `"principalId":"` + testEventGridSP, `"value":"unrelated"`} {
					if !strings.Contains(string(calls), expected) {
						t.Errorf("missing %q from calls: %s", expected, calls)
					}
				}
			} else if strings.Contains(string(calls), `"method":"POST"`) || strings.Contains(string(calls), `"method":"PATCH"`) {
				t.Fatalf("existing setup mutated objects: %s", calls)
			}
		})
	}
}

// Executed as a separate process by the mock az command.
func TestGraphCLIHelper(t *testing.T) {
	if os.Getenv("GRAPH_HELPER") != "1" {
		return
	}
	args := os.Args
	for len(args) > 0 && args[0] != "--" {
		args = args[1:]
	}
	args = args[1:]
	mode := os.Getenv("GRAPH_MODE")
	fail := func() {
		fmt.Fprintln(os.Stderr, "Mock Graph request failed")
		os.Exit(42)
	}
	emit := func(value any) {
		if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	if args[0] == "account" {
		if mode == "no-login" {
			fail()
		}
		identity := "user"
		if mode == "non-user" {
			identity = "servicePrincipal"
		}
		emit(map[string]any{"user": map[string]string{"type": identity}})
	}
	values := map[string]string{}
	for i := 1; i+1 < len(args); i += 2 {
		values[args[i]] = args[i+1]
	}
	uri, err := url.Parse(values["--url"])
	if err != nil || uri.Host != "graph.microsoft.com" {
		fail()
	}
	method := values["--method"]
	var body map[string]any
	if values["--body"] != "" {
		if json.Unmarshal([]byte(values["--body"]), &body) != nil {
			fail()
		}
	}
	record, _ := json.Marshal(map[string]any{"method": method, "url": uri.String(), "body": body})
	f, err := os.OpenFile(os.Getenv("GRAPH_CALLS"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		fail()
	}
	_, _ = fmt.Fprintln(f, string(record))
	_ = f.Close()
	role := map[string]any{"id": testRoleID, "value": "AzureEventGridSecureWebhookSubscriber", "isEnabled": true, "allowedMemberTypes": []string{"User", "Application"}}
	roles := []any{map[string]any{"id": testObjectID, "value": "unrelated"}}
	existing := mode == "existing" || mode == "paginated"
	if existing {
		roles = append(roles, role)
	}
	app := map[string]any{"id": testObjectID, "appId": testAppID, "appRoles": roles}
	switch {
	case uri.Path == "/v1.0/me":
		emit(map[string]string{"id": testUserID})
	case uri.Path == "/v1.0/applications" && method == "POST":
		if mode == "missing-id" {
			delete(app, "id")
		}
		emit(app)
	case uri.Path == "/v1.0/applications" && method == "GET":
		if uri.Query().Get("$filter") != "appId eq '"+testAppID+"'" {
			fail()
		}
		if mode == "unknown-app" {
			emit(map[string]any{"value": []any{}})
		}
		emit(map[string]any{"value": []any{app}})
	case uri.Path == "/v1.0/applications/"+testObjectID && method == "PATCH":
		if mode == "patch-failure" {
			fail()
		}
		items, ok := body["appRoles"].([]any)
		if !ok || len(items) != 2 {
			fail()
		}
		emit(nil)
	case uri.Path == "/v1.0/servicePrincipals" && method == "GET":
		if mode == "forbidden" {
			fail()
		}
		if mode == "malformed-lookup" {
			emit(map[string]any{"value": nil})
		}
		filter := uri.Query().Get("$filter")
		id := testWebhookSP
		if filter == "appId eq '4962773b-9cdb-44cf-a8bf-237846a00ab7'" {
			id = testEventGridSP
		} else if filter != "appId eq '"+testAppID+"'" {
			fail()
		}
		items := []any{}
		if existing {
			items = append(items, map[string]string{"id": id})
		}
		emit(map[string]any{"value": items})
	case uri.Path == "/v1.0/servicePrincipals" && method == "POST":
		if mode == "create-failure" {
			fail()
		}
		if mode == "principal-missing-id" {
			emit(map[string]any{})
		}
		id := testWebhookSP
		if body["appId"] == "4962773b-9cdb-44cf-a8bf-237846a00ab7" {
			id = testEventGridSP
		} else if body["appId"] != testAppID {
			fail()
		}
		emit(map[string]string{"id": id})
	case uri.Path == "/v1.0/servicePrincipals/"+testWebhookSP:
		items := []any{role}
		if mode == "propagation" {
			calls, _ := os.ReadFile(os.Getenv("GRAPH_CALLS"))
			if strings.Count(string(calls), `"url":"https://graph.microsoft.com/v1.0/servicePrincipals/`+testWebhookSP+`"`) < 2 {
				items = []any{}
			}
		}
		if mode == "no-role" {
			items = nil
		}
		emit(map[string]any{"id": testWebhookSP, "appRoles": items})
	case uri.Path == "/v1.0/servicePrincipals/"+testWebhookSP+"/appRoleAssignedTo":
		if method == "POST" {
			if mode == "assignment-failure" {
				fail()
			}
			if body["resourceId"] != testWebhookSP || body["appRoleId"] != testRoleID || (body["principalId"] != testUserID && body["principalId"] != testEventGridSP) {
				fail()
			}
			emit(map[string]string{"id": "opaque-role-assignment-id"})
		}
		if mode == "paginated" && uri.Query().Get("page") == "" {
			emit(map[string]any{"value": []any{}, "@odata.nextLink": "https://graph.microsoft.com" + uri.Path + "?page=2"})
		}
		items := []any{}
		if existing {
			for _, principal := range []string{testUserID, testEventGridSP} {
				items = append(items, map[string]string{"principalId": principal, "resourceId": testWebhookSP, "appRoleId": testRoleID})
			}
		}
		emit(map[string]any{"value": items})
	default:
		fail()
	}
}
