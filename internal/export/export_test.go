package export

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/wizli595/tidydir/internal/action"
)

func TestToJSON(t *testing.T) {
	actions := []action.Action{
		{Type: action.ActionMove, Source: "/tmp/a.txt", Dest: "/tmp/docs/a.txt", Reason: "document"},
		{Type: action.ActionDelete, Source: "/tmp/junk.tmp", Reason: "junk file"},
	}

	var buf bytes.Buffer
	if err := ToJSON(&buf, actions); err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var entries []PlanEntry
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Type != "move" {
		t.Errorf("expected type 'move', got %q", entries[0].Type)
	}
	if entries[1].Type != "delete" {
		t.Errorf("expected type 'delete', got %q", entries[1].Type)
	}
}

func TestToCSV(t *testing.T) {
	actions := []action.Action{
		{Type: action.ActionMove, Source: "/tmp/a.txt", Dest: "/tmp/docs/a.txt", Reason: "document"},
		{Type: action.ActionRename, Source: "/tmp/My File.txt", Dest: "/tmp/my-file.txt", Reason: "normalize"},
	}

	var buf bytes.Buffer
	if err := ToCSV(&buf, actions); err != nil {
		t.Fatalf("ToCSV failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "type,source,dest,reason") {
		t.Errorf("unexpected header: %s", lines[0])
	}
}

func TestToJSONEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := ToJSON(&buf, []action.Action{}); err != nil {
		t.Fatalf("ToJSON failed on empty: %v", err)
	}

	var entries []PlanEntry
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}
