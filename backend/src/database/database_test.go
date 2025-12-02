package database

import "testing"

func TestFirstNonEmpty(t *testing.T) {
	got := firstNonEmpty("", "foo", "bar")
	if got != "foo" {
		t.Fatalf("firstNonEmpty returned %q, want %q", got, "foo")
	}
	if firstNonEmpty("", "", "") != "" {
		t.Fatalf("firstNonEmpty should return empty string when all inputs are empty")
	}
}
