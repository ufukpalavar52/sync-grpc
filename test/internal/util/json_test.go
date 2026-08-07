package util_test

import (
	"encoding/json"
	"imapsync-user/internal/util"
	"testing"
)

func TestJsonEncode(t *testing.T) {
	data := map[string]any{"foo": "bar", "count": float64(3)}
	encoded := util.JsonEncode(data)

	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("expected valid json, got error: %v", err)
	}
	if decoded["foo"] != "bar" {
		t.Fatalf("expected foo=bar, got %v", decoded["foo"])
	}
	if decoded["count"] != float64(3) {
		t.Fatalf("expected count=3, got %v", decoded["count"])
	}
}

func TestJsonEncodeStr(t *testing.T) {
	encoded := util.JsonEncodeStr([]string{"a", "b", "c"})
	if encoded != `["a","b","c"]` {
		t.Fatalf("expected %q, got %q", `["a","b","c"]`, encoded)
	}
}

func TestJsonEncode_UnmarshalableReturnsNil(t *testing.T) {
	encoded := util.JsonEncode(make(chan int))
	if encoded != nil {
		t.Fatalf("expected nil for unmarshalable input, got %v", encoded)
	}
}
