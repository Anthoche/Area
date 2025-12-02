package workflows

import "testing"

func TestIntervalConfigFromJSON(t *testing.T) {
	raw := []byte(`{"interval_minutes":5,"payload":{"foo":"bar"}}`)
	cfg, err := intervalConfigFromJSON(raw)
	if err != nil {
		t.Fatalf("intervalConfigFromJSON returned error: %v", err)
	}
	if cfg.IntervalMinutes != 5 {
		t.Fatalf("interval minutes = %d, want 5", cfg.IntervalMinutes)
	}
	if cfg.Payload["foo"] != "bar" {
		t.Fatalf("payload foo = %v, want bar", cfg.Payload["foo"])
	}
}

func TestIntervalConfigFromJSON_Invalid(t *testing.T) {
	_, err := intervalConfigFromJSON([]byte(`not json`))
	if err == nil {
		t.Fatalf("expected error for invalid json")
	}
}
