package wasp

import (
	"encoding/binary"
	"fmt"
)

type Iterator struct {
	data []byte
	size int
	pos  int
}

func (it *Iterator) uint32() uint32 {
	value := binary.LittleEndian.Uint32(it.data[it.pos:])
	it.pos += 4
	return value
}

func (it *Iterator) done() bool {
	return it.pos >= it.size
}

func (it *Iterator) byte() byte {
	b := it.data[it.pos]
	it.pos++
	return b
}

func (it *Iterator) bytes(n int) []byte {
	b := it.data[it.pos : it.pos+n]
	it.pos += n
	return b
}

func (it *Iterator) readUntil(target byte) ([]byte, error) {
	var bytes []byte
	for !it.done() {
		b := it.byte()
		bytes = append(bytes, b)
		if b == target {
			return bytes, nil
		}
	}
	return nil, fmt.Errorf("target byte 0x%x not found", target)
}

// All integer constants are encoded using a space-efficient, variable-length LEB128 encoding
func (it *Iterator) varint() int {
	var x uint64

	b := it.data[it.pos]
	x = uint64(b & 0x7F)
	if b < 0x80 {
		it.pos++
		return int(x)
	}

	panic("varint parsing not implemented yet")
}

func newIterator(data []byte) *Iterator {
	return &Iterator{
		data: data,
		size: len(data),
		pos:  0,
	}
}
