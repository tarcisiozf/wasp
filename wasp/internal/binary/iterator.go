package binary

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

	b = it.data[it.pos+1]
	x |= uint64(b&0x7F) << 7
	if b < 0x80 {
		it.pos += 2
		return int(x)
	}

	b = it.data[it.pos+2]
	x |= uint64(b&0x7F) << 14
	if b < 0x80 {
		it.pos += 3
		return int(x)
	}

	b = it.data[it.pos+3]
	x |= uint64(b&0x7F) << 21
	if b < 0x80 {
		it.pos += 4
		return int(x)
	}

	panic("varint parsing not implemented yet")
}

func (it *Iterator) String(len int) string {
	return string(it.Bytes(len))
}

func (it *Iterator) Peek() byte {
	return it.data[it.pos]
}

func (it *Iterator) PeekAt(n int) byte {
	return it.data[it.pos+n]
}

func (it *Iterator) Next() {
	it.pos++
}

func (it *Iterator) BoolByte() bool {
	value := it.Byte()
	return value != 0
}

func (it *Iterator) Assert(expected ...byte) {
	bytes := it.Bytes(len(expected))
	for i, b := range expected {
		if bytes[i] != b {
			panic(fmt.Sprintf("assertion failed: expected bytes %v, got %v", expected, bytes))
		}
	}
}

func (it *Iterator) Position() int {
	return it.pos
}

func NewIterator(data []byte) *Iterator {
	return &Iterator{
		data: data,
		size: len(data),
		pos:  0,
	}
}
