package main

import (
	"io/fs"
	"strings"
	"testing"
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
