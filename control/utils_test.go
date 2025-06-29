package control

import "testing"

func TestContains(t *testing.T) {
	data := "nil"
	sData := []string{
		"no",
		"nil",
	}
	invalidData := 0
	result := contains(data, "true")
	if result != false {
		t.Fatalf("Expected contains(%q, %q) to be false, but got true", data, "true")
	}
	result = contains(data, "nil")
	if result != true {
		t.Fatalf("Expected contains(%q, %q) to be true, but got false", data, "nil")
	}
	result = contains(sData, "true")
	if result != false {
		t.Fatalf("Expected contains(%v, %q) to be false, but got true", sData, "true")
	}
	result = contains(sData, "nil")
	if result != true {
		t.Fatalf("Expected contains(%v, %q) to be true, but got false", sData, "nil")
	}
	result = contains(invalidData, "hoge")
	if result != false {
		t.Fatalf("Expected contains(%v, %q) to be false for invalid data, but got true", invalidData, "hoge")
	}
}
