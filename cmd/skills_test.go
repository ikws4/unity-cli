package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"
)

func testSkillFS() fstest.MapFS {
	return fstest.MapFS{
		"unity-cli/SKILL.md": {
			Data: []byte("---\nname: unity-cli\nversion: 1.0.0\ndescription: \"Unity guidance\"\n---\n# Body\n"),
		},
		"unity-cli/references/exec.md": {Data: []byte("# Exec reference\n")},
	}
}

func TestRunSkillsCmdList(t *testing.T) {
	var out strings.Builder
	if err := runSkillsCmd(testSkillFS(), []string{"list"}, &out); err != nil {
		t.Fatalf("runSkillsCmd: %v", err)
	}

	var got struct {
		Skills []skillInfo `json:"skills"`
		Count  int         `json:"count"`
	}
	if err := json.Unmarshal([]byte(out.String()), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out.String())
	}
	if got.Count != 1 || len(got.Skills) != 1 {
		t.Fatalf("skills: %+v", got)
	}
	if got.Skills[0].Name != "unity-cli" || got.Skills[0].Version != "1.0.0" || got.Skills[0].Description != "Unity guidance" {
		t.Fatalf("skill info: %+v", got.Skills[0])
	}
}

func TestRunSkillsCmdListPath(t *testing.T) {
	var out strings.Builder
	if err := runSkillsCmd(testSkillFS(), []string{"list", "unity-cli"}, &out); err != nil {
		t.Fatalf("runSkillsCmd: %v", err)
	}

	var got struct {
		Path    string       `json:"path"`
		Entries []skillEntry `json:"entries"`
		Count   int          `json:"count"`
	}
	if err := json.Unmarshal([]byte(out.String()), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out.String())
	}
	if got.Path != "unity-cli" || got.Count != 2 {
		t.Fatalf("listing: %+v", got)
	}
	if got.Entries[0].Path != "unity-cli/SKILL.md" || got.Entries[0].IsDir {
		t.Fatalf("first entry: %+v", got.Entries[0])
	}
	if got.Entries[1].Path != "unity-cli/references" || !got.Entries[1].IsDir {
		t.Fatalf("second entry: %+v", got.Entries[1])
	}
}

func TestRunSkillsCmdRead(t *testing.T) {
	for _, args := range [][]string{
		{"read", "unity-cli"},
		{"read", "unity-cli/SKILL.md"},
	} {
		var out strings.Builder
		if err := runSkillsCmd(testSkillFS(), args, &out); err != nil {
			t.Fatalf("runSkillsCmd(%v): %v", args, err)
		}
		if !strings.Contains(out.String(), "# Body") {
			t.Fatalf("content for %v: %q", args, out.String())
		}
	}

	var reference strings.Builder
	if err := runSkillsCmd(testSkillFS(), []string{"read", "unity-cli/references/exec.md"}, &reference); err != nil {
		t.Fatalf("read reference: %v", err)
	}
	if reference.String() != "# Exec reference\n" {
		t.Fatalf("reference content: %q", reference.String())
	}
}

func TestRunSkillsCmdRejectsPathTraversal(t *testing.T) {
	for _, target := range []string{"../secret", "unity-cli/../../secret", `unity-cli\..\secret`, "/etc/passwd"} {
		var out strings.Builder
		err := runSkillsCmd(testSkillFS(), []string{"read", target}, &out)
		if err == nil {
			t.Fatalf("expected %q to be rejected", target)
		}
		if out.Len() != 0 {
			t.Fatalf("output leaked for %q: %q", target, out.String())
		}
	}
}

func TestRunSkillsCmdRequiresEmbeddedContent(t *testing.T) {
	var out strings.Builder
	err := runSkillsCmd(nil, []string{"list"}, &out)
	if err == nil || !strings.Contains(err.Error(), "not embedded") {
		t.Fatalf("expected missing embed error, got %v", err)
	}
}
