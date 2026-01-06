package database

import (
	"area/src/database"
	"testing"
)

func TestFirstNonEmpty(t *testing.T) {
	got := database.FirstNonEmpty("", "foo", "bar")
	if got != "foo" {
		t.Fatalf("firstNonEmpty returned %q, want %q", got, "foo")
	}
	if database.FirstNonEmpty("", "", "") != "" {
		t.Fatalf("firstNonEmpty should return empty string when all inputs are empty")
	}
}
