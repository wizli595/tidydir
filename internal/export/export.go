package export

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"path/filepath"

	"github.com/wizli595/tidydir/internal/action"
)

// PlanEntry is a JSON-friendly representation of an action.
type PlanEntry struct {
	Type   string `json:"type"`
	Source string `json:"source"`
	Dest   string `json:"dest,omitempty"`
	Reason string `json:"reason"`
}

// ToJSON writes the action plan as indented JSON.
func ToJSON(w io.Writer, actions []action.Action) error {
	entries := make([]PlanEntry, len(actions))
	for i, a := range actions {
		entries[i] = PlanEntry{
			Type:   string(a.Type),
			Source: a.Source,
			Dest:   a.Dest,
			Reason: a.Reason,
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

// ToCSV writes the action plan as CSV with a header row.
func ToCSV(w io.Writer, actions []action.Action) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	if err := writer.Write([]string{"type", "source", "dest", "reason"}); err != nil {
		return err
	}
	for _, a := range actions {
		dest := ""
		if a.Dest != "" {
			dest = filepath.Base(a.Dest)
		}
		if err := writer.Write([]string{
			string(a.Type),
			filepath.Base(a.Source),
			dest,
			a.Reason,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}
