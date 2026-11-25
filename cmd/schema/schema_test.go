package schema

import "testing"

func TestDefinitions(t *testing.T) {
	for _, name := range []string{"context", "structure", "ask", "impact", "repo-map", "pack", "doctor", "fix", "analyze", "agent-bench"} {
		d, ok := Get(name)
		if !ok {
			t.Fatalf("missing schema for %s", name)
		}
		if d.SchemaVersion != 1 || len(d.Fields) == 0 {
			t.Fatalf("incomplete schema for %s: %+v", name, d)
		}
	}
	if len(List()) < 10 {
		t.Fatalf("expected at least 10 published schemas, got %d", len(List()))
	}
	if _, ok := Get("missing"); ok {
		t.Fatal("unexpected schema for unknown command")
	}
}
