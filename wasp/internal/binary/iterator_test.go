package binary_test

import (
	"testing"
	"wasp/wasp/internal/binary"
)

func TestIterator_Varint(t *testing.T) {
	iter := binary.NewIterator([]byte{12})
	value := iter.Varint()
	if value != 12 {
		t.Fatalf("expected 12, got %d", value)
	}
}
