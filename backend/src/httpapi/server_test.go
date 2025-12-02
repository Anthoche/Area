package httpapi

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestEnsureNoTrailingData_OK(t *testing.T) {
	dec := json.NewDecoder(strings.NewReader(`{"a":1}`))
	dec.DisallowUnknownFields()
	var payload map[string]any
	if err := dec.Decode(&payload); err != nil {
		t.Fatalf("decode first object: %v", err)
	}
	if err := ensureNoTrailingData(dec); err != nil {
		t.Fatalf("ensureNoTrailingData returned error: %v", err)
	}
}

func TestEnsureNoTrailingData_ExtraObject(t *testing.T) {
	buf := bytes.NewBufferString(`{"a":1} {"b":2}`)
	dec := json.NewDecoder(buf)
	dec.DisallowUnknownFields()
	var payload map[string]any
	if err := dec.Decode(&payload); err != nil {
		t.Fatalf("decode first object: %v", err)
	}
	if err := ensureNoTrailingData(dec); err == nil {
		t.Fatalf("expected error for trailing object, got nil")
	}
}
