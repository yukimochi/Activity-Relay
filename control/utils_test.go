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
		t.Fatalf("fail - not contain but return true")
	}
	result = contains(data, "nil")
	if result != true {
		t.Fatalf("fail - contains but return false")
	}
	result = contains(sData, "true")
	if result != false {
		t.Fatalf("fail - not contain but return true (slice)")
	}
	result = contains(sData, "nil")
	if result != true {
		t.Fatalf("fail - contains but return false (slice)")
	}
	result = contains(invalidData, "hoge")
	if result != false {
		t.Fatalf("fail - given invalid data but return true (slice)")
	}
}
