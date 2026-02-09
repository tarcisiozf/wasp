package iterator

import "testing"

func TestIterator_Varint(t *testing.T) {
	iter := NewIterator([]byte{12})
	value := iter.Varint()
	if value != 12 {
		t.Fatalf("expected 12, got %d", value)
	}
}
