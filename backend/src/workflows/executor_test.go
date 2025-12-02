package workflows

import "testing"

func TestDecodePayload(t *testing.T) {
	payload := []byte(`{"foo":"bar","n":3}`)
	var target struct {
		Foo string `json:"foo"`
		N   int    `json:"n"`
	}

	if err := DecodePayload(payload, &target); err != nil {
		t.Fatalf("DecodePayload error: %v", err)
	}
	if target.Foo != "bar" || target.N != 3 {
		t.Fatalf("decoded struct mismatch: %+v", target)
	}
}

func TestDecodePayload_Empty(t *testing.T) {
	var target struct {
		Value string `json:"value"`
	}
	if err := DecodePayload([]byte{}, &target); err != nil {
		t.Fatalf("DecodePayload should succeed on empty payload: %v", err)
	}
	if target.Value != "" {
		t.Fatalf("target should remain zero-value, got %q", target.Value)
	}
}
