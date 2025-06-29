package control

import "testing"

func TestContainsFunction(t *testing.T) {
	t.Run("String data contains target", func(t *testing.T) {
		data := "nil"
		result := contains(data, "nil")
		if result != true {
			t.Fatalf("Expected contains(%q, %q) to be true, but got false", data, "nil")
		}
	})

	t.Run("String data does not contain target", func(t *testing.T) {
		data := "nil"
		result := contains(data, "true")
		if result != false {
			t.Fatalf("Expected contains(%q, %q) to be false, but got true", data, "true")
		}
	})

	t.Run("String slice contains target", func(t *testing.T) {
		sData := []string{"no", "nil"}
		result := contains(sData, "nil")
		if result != true {
			t.Fatalf("Expected contains(%v, %q) to be true, but got false", sData, "nil")
		}
	})

	t.Run("String slice does not contain target", func(t *testing.T) {
		sData := []string{"no", "nil"}
		result := contains(sData, "true")
		if result != false {
			t.Fatalf("Expected contains(%v, %q) to be false, but got true", sData, "true")
		}
	})

	t.Run("Invalid data type returns false", func(t *testing.T) {
		invalidData := 0
		result := contains(invalidData, "hoge")
		if result != false {
			t.Fatalf("Expected contains(%v, %q) to be false for invalid data, but got true", invalidData, "hoge")
		}
	})
}
