package iterator

import (
	"encoding/binary"
	"fmt"
)

type Iterator struct {
	data []byte
	size int
	pos  int
}

func (it *Iterator) Uint32() uint32 {
	value := binary.LittleEndian.Uint32(it.data[it.pos:])
	it.pos += 4
	return value
}

func (it *Iterator) Done() bool {
	return it.pos >= it.size
}

func (it *Iterator) Byte() byte {
	b := it.data[it.pos]
	it.pos++
	return b
}

func (it *Iterator) Bytes(n int) []byte {
	b := it.data[it.pos : it.pos+n]
	it.pos += n
	return b
}

func (it *Iterator) ReadUntil(target byte) ([]byte, error) {
	var bytes []byte
	for !it.Done() {
		b := it.Byte()
		bytes = append(bytes, b)
		if b == target {
			return bytes, nil
		}
	}
	return nil, fmt.Errorf("target byte 0x%x not found", target)
}

// All integer constants are encoded using a space-efficient, variable-length LEB128 encoding
func (it *Iterator) Varint() int {
	var x uint64

	b := it.data[it.pos]
	x = uint64(b & 0x7F)
	if b < 0x80 {
		it.pos++
		return int(x)
	}

	panic("varint parsing not implemented yet")
}

func (it *Iterator) String(len int) string {
	return string(it.Bytes(len))
}

func NewIterator(data []byte) *Iterator {
	return &Iterator{
		data: data,
		size: len(data),
		pos:  0,
	}
}
