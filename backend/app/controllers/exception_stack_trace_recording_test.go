package controllers

import (
	"encoding/json"
	"testing"
)

// Legacy S3 files stored only the events array as their top-level JSON. The
// new shape is a wrapper object with `events` (and optionally `logs`,
// `actions`, `startedAt`, `endedAt`). The frontend should always see the
// wrapper, regardless of when the recording was written.
func TestWrapLegacyRecording(t *testing.T) {
	cases := []struct {
		name string
		in   string
		// We compare by parsing both sides as JSON to avoid depending on
		// formatting/whitespace.
		wantEvents string
		wantWrap   bool
	}{
		{
			name:       "legacy events-only array",
			in:         `[{"type":"flutter_video","fps":15}]`,
			wantEvents: `[{"type":"flutter_video","fps":15}]`,
			wantWrap:   true,
		},
		{
			name:       "legacy with leading whitespace",
			in:         "\n  [1,2,3]",
			wantEvents: `[1,2,3]`,
			wantWrap:   true,
		},
		{
			name:       "new wrapper object passed through",
			in:         `{"events":[1],"logs":[],"actions":[]}`,
			wantEvents: `[1]`,
			wantWrap:   false,
		},
		{
			name:       "empty new wrapper",
			in:         `{}`,
			wantEvents: ``, // no events field
			wantWrap:   false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := wrapLegacyRecording([]byte(c.in))
			var parsed map[string]json.RawMessage
			if err := json.Unmarshal(got, &parsed); err != nil {
				t.Fatalf("output not a JSON object: %v\noutput=%s", err, string(got))
			}
			events, ok := parsed["events"]
			if c.wantEvents == "" {
				if ok {
					t.Fatalf("expected no events field, got %s", string(events))
				}
				return
			}
			if !ok {
				t.Fatalf("missing events field; got %s", string(got))
			}
			// Normalize both sides through JSON to ignore whitespace differences.
			var gotEvents, wantEvents any
			if err := json.Unmarshal(events, &gotEvents); err != nil {
				t.Fatalf("events not valid JSON: %v", err)
			}
			if err := json.Unmarshal([]byte(c.wantEvents), &wantEvents); err != nil {
				t.Fatalf("wantEvents not valid JSON: %v", err)
			}
			if !jsonEqual(gotEvents, wantEvents) {
				t.Fatalf("events mismatch:\n got=%v\nwant=%v", gotEvents, wantEvents)
			}
		})
	}
}

func jsonEqual(a, b any) bool {
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	return string(ja) == string(jb)
}
