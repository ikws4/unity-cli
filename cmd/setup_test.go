package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createUnityProject(t *testing.T, manifest string) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{"Assets", "Packages", "ProjectSettings"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "ProjectSettings", "ProjectVersion.txt"), []byte("m_EditorVersion: 2022.3.0f1\n"), 0o644); err != nil {
		t.Fatalf("write ProjectVersion.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "Packages", "manifest.json"), []byte(manifest), 0o640); err != nil {
		t.Fatalf("write manifest.json: %v", err)
	}
	return root
}

func readManifest(t *testing.T, root string) map[string]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "Packages", "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest.json: %v", err)
	}
	var manifest map[string]json.RawMessage
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse manifest.json: %v\n%s", err, data)
	}
	return manifest
}

func manifestDependencies(t *testing.T, manifest map[string]json.RawMessage) map[string]string {
	t.Helper()
	var dependencies map[string]string
	if err := json.Unmarshal(manifest["dependencies"], &dependencies); err != nil {
		t.Fatalf("parse dependencies: %v", err)
	}
	return dependencies
}

func TestRunSetupInstallsVersionedConnector(t *testing.T) {
	root := createUnityProject(t, `{
  "dependencies": {
    "com.unity.test-framework": "1.1.33"
  },
  "testables": ["com.example.package"]
}`)

	var out strings.Builder
	if err := runSetup(root, "v0.3.22", &out); err != nil {
		t.Fatalf("runSetup: %v", err)
	}

	manifest := readManifest(t, root)
	dependencies := manifestDependencies(t, manifest)
	wantURL := connectorPackageURL + "#v0.3.22"
	if dependencies[connectorPackageName] != wantURL {
		t.Fatalf("connector dependency = %q, want %q", dependencies[connectorPackageName], wantURL)
	}
	if dependencies["com.unity.test-framework"] != "1.1.33" {
		t.Fatalf("existing dependency was changed: %+v", dependencies)
	}
	var testables []string
	if err := json.Unmarshal(manifest["testables"], &testables); err != nil {
		t.Fatalf("parse preserved testables: %v", err)
	}
	if len(testables) != 1 || testables[0] != "com.example.package" {
		t.Fatalf("unknown manifest field was changed: %+v", testables)
	}
	if !strings.Contains(out.String(), "Installed Unity CLI Connector") {
		t.Fatalf("output = %q", out.String())
	}

	info, err := os.Stat(filepath.Join(root, "Packages", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("manifest mode = %o, want 640", info.Mode().Perm())
	}
}

func TestRunSetupUpdatesExistingConnector(t *testing.T) {
	root := createUnityProject(t, `{"dependencies":{"com.ikws4.unity-cli-connector":"old"}}`)

	var out strings.Builder
	if err := runSetup(root, "0.4.0", &out); err != nil {
		t.Fatalf("runSetup: %v", err)
	}

	dependencies := manifestDependencies(t, readManifest(t, root))
	if got, want := dependencies[connectorPackageName], connectorPackageURL+"#v0.4.0"; got != want {
		t.Fatalf("connector dependency = %q, want %q", got, want)
	}
	if !strings.Contains(out.String(), "Updated Unity CLI Connector") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunSetupIsIdempotent(t *testing.T) {
	wantURL := connectorPackageURL + "#v0.3.22"
	root := createUnityProject(t, `{"dependencies":{"com.ikws4.unity-cli-connector":"`+wantURL+`"}}`)
	manifestPath := filepath.Join(root, "Packages", "manifest.json")
	before, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	if err := runSetup(root, "v0.3.22", &out); err != nil {
		t.Fatalf("runSetup: %v", err)
	}
	after, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("idempotent setup rewrote manifest:\nbefore: %s\nafter: %s", before, after)
	}
	if !strings.Contains(out.String(), "already configured") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunSetupMigratesLegacyConnector(t *testing.T) {
	root := createUnityProject(t, `{"dependencies":{"com.youngwoocho02.unity-cli-connector":"old"}}`)

	var out strings.Builder
	if err := runSetup(root, "v0.4.0", &out); err != nil {
		t.Fatalf("runSetup: %v", err)
	}

	dependencies := manifestDependencies(t, readManifest(t, root))
	if _, exists := dependencies[legacyConnectorPackageName]; exists {
		t.Fatalf("legacy connector dependency was not removed: %+v", dependencies)
	}
	if got, want := dependencies[connectorPackageName], connectorPackageURL+"#v0.4.0"; got != want {
		t.Fatalf("connector dependency = %q, want %q", got, want)
	}
	if !strings.Contains(out.String(), "Updated Unity CLI Connector") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunSetupUsesDefaultBranchForDevelopmentBuild(t *testing.T) {
	root := createUnityProject(t, `{}`)

	var out strings.Builder
	if err := runSetup(root, "dev", &out); err != nil {
		t.Fatalf("runSetup: %v", err)
	}

	dependencies := manifestDependencies(t, readManifest(t, root))
	if got := dependencies[connectorPackageName]; got != connectorPackageURL {
		t.Fatalf("connector dependency = %q, want %q", got, connectorPackageURL)
	}
}

func TestRunSetupRequiresUnityProjectRoot(t *testing.T) {
	var out strings.Builder
	err := runSetup(t.TempDir(), "dev", &out)
	if err == nil || !strings.Contains(err.Error(), "Unity project root") {
		t.Fatalf("expected project root error, got %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRunSetupRejectsInvalidManifest(t *testing.T) {
	root := createUnityProject(t, `{invalid`)

	var out strings.Builder
	err := runSetup(root, "dev", &out)
	if err == nil || !strings.Contains(err.Error(), "parse Packages/manifest.json") {
		t.Fatalf("expected manifest parse error, got %v", err)
	}
}

func TestSetupCmdRejectsArguments(t *testing.T) {
	err := setupCmd([]string{"/another/project"})
	if err == nil || !strings.Contains(err.Error(), "does not accept arguments") {
		t.Fatalf("expected argument error, got %v", err)
	}
}
