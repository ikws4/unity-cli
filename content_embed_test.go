package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/ikws4/unity-cli/cmd"
)

func TestUnityCliSkillIsEmbedded(t *testing.T) {
	data, err := fs.ReadFile(embeddedContent, "skills/unity-cli/SKILL.md")
	if err != nil {
		t.Fatalf("read embedded skill: %v", err)
	}
	if !strings.Contains(string(data), "UnityEngine.Object") {
		t.Fatal("embedded skill is missing exec Object guidance")
	}
}

func TestDefaultVersionMatchesConnector(t *testing.T) {
	if Version != cmd.DefaultVersion {
		t.Fatalf("main Version = %q, cmd DefaultVersion = %q", Version, cmd.DefaultVersion)
	}

	packageData, err := os.ReadFile("unity-connector/package.json")
	if err != nil {
		t.Fatalf("read connector package: %v", err)
	}
	var connectorPackage struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(packageData, &connectorPackage); err != nil {
		t.Fatalf("parse connector package: %v", err)
	}
	if connectorPackage.Version != cmd.DefaultVersion {
		t.Fatalf("connector package version = %q, want %q", connectorPackage.Version, cmd.DefaultVersion)
	}

	heartbeat, err := os.ReadFile("unity-connector/Editor/Heartbeat.cs")
	if err != nil {
		t.Fatalf("read connector heartbeat: %v", err)
	}
	wantHeartbeatVersion := `const string CONNECTOR_VERSION = "` + cmd.DefaultVersion + `";`
	if !strings.Contains(string(heartbeat), wantHeartbeatVersion) {
		t.Fatalf("connector heartbeat does not contain %q", wantHeartbeatVersion)
	}
}
