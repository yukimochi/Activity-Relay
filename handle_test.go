package main

import "testing"

func TestContains(t *testing.T) {
	data := "nil"
	sData := []string{
		"no",
		"nil",
	}
	badData := 0
	result := contains(data, "true")
	if result != false {
		t.Fatalf("Failed - no contain but true.")
	}
	result = contains(data, "nil")
	if result != true {
		t.Fatalf("Failed - contain but false.")
	}
	result = contains(sData, "true")
	if result != false {
		t.Fatalf("Failed - no contain but true. (slice)")
	}
	result = contains(sData, "nil")
	if result != true {
		t.Fatalf("Failed - contain but false. (slice)")
	}
	result = contains(badData, "hoge")
	if result != false {
		t.Fatalf("Failed - input bad data but true. (slice)")
	}
}
