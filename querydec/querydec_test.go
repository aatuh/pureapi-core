package querydec

import (
	"net/url"
	"reflect"
	"testing"
)

func TestPlainDecoder_Decode(t *testing.T) {
	decoder := PlainDecoder{}

	values := url.Values{
		"x": []string{"1"},
		"y": []string{"a"},
		"z": []string{"b", "c"},
	}

	result, err := decoder.Decode(values)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := map[string]any{
		"x": "1",
		"y": "a",
		"z": []string{"b", "c"},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Expected %v, got %v", expected, result)
	}
}

func TestPlainDecoder_Decode_Empty(t *testing.T) {
	decoder := PlainDecoder{}

	values := url.Values{}

	result, err := decoder.Decode(values)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result) != 0 {
		t.Fatalf("Expected empty result, got %v", result)
	}
}
