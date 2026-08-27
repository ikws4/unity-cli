package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ikws4/unity-cli/internal/client"
)

func TestTestCmd_SendsDefaultRunParameters(t *testing.T) {
	var captured map[string]interface{}
	send := func(cmd string, params interface{}) (*client.CommandResponse, error) {
		if cmd != "run_tests" {
			t.Fatalf("send called with command %q, want run_tests", cmd)
		}
		var ok bool
		captured, ok = params.(map[string]interface{})
		if !ok {
			t.Fatalf("params type = %T, want map[string]interface{}", params)
		}
		return &client.CommandResponse{Success: true}, nil
	}

	resp, err := testCmd([]string{}, send, nil)
	if err != nil {
		t.Fatalf("testCmd returned error: %v", err)
	}
	if resp == nil || !resp.Success {
		t.Fatalf("testCmd response = %#v, want success", resp)
	}
	if captured["mode"] != "EditMode" {
		t.Errorf("mode = %v, want EditMode", captured["mode"])
	}
	if captured["runId"] == "" {
		t.Error("runId should be sent")
	}
}

func TestPollTestResultsStopsWhenProjectInstanceDisappears(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	origPollInterval := statusPollInterval
	statusPollInterval = time.Millisecond
	t.Cleanup(func() { statusPollInterval = origPollInterval })

	_, err := pollTestResults("missing", func() (*client.Instance, error) {
		return nil, fmt.Errorf("no Unity instance found for project: /projects/current")
	})
	if err == nil {
		t.Fatal("expected stopped editor error")
	}
	if !strings.Contains(err.Error(), "unity editor has stopped") {
		t.Fatalf("expected stopped editor error, got %v", err)
	}
}

func TestTestCmd_PlayModePollsRunIDResult(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	statusDir := filepath.Join(home, ".unity-cli", "status")
	if err := os.MkdirAll(statusDir, 0755); err != nil {
		t.Fatalf("failed to create status dir: %v", err)
	}

	send := func(cmd string, params interface{}) (*client.CommandResponse, error) {
		captured := params.(map[string]interface{})
		runID := captured["runId"].(string)
		resp := client.CommandResponse{Success: true, Message: "done"}
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("failed to marshal response: %v", err)
		}
		if err := os.WriteFile(filepath.Join(statusDir, "test-results-"+runID+".json"), data, 0644); err != nil {
			t.Fatalf("failed to write results: %v", err)
		}
		return &client.CommandResponse{Success: true, Message: "running"}, nil
	}

	resp, err := testCmd([]string{"--mode", "PlayMode"}, send, nil)
	if err != nil {
		t.Fatalf("testCmd returned error: %v", err)
	}
	if resp.Message != "done" {
		t.Fatalf("Message = %q, want done", resp.Message)
	}
}
