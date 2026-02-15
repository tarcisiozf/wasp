package binary_test

import (
	"testing"
	"wasp/wasp/internal/binary"
)

func TestIterator_Varint(t *testing.T) {
	t.Run("single byte", func(t *testing.T) {
		iter := binary.NewIterator([]byte{12})
		value := iter.Varint()
		if value != 12 {
			t.Fatalf("expected 12, got %d", value)
		}
	})
	t.Run("i64 literal", func(t *testing.T) {
		iter := binary.NewIterator([]byte{
			0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x7f,
		})
		value := uint64(iter.Varint())
		if value != 9223372036854775808 {
			t.Fatalf("expected 9223372036854775808, got %d", value)
		}
	})
}
