package binary

import (
	"encoding/binary"
	"fmt"
	"unsafe"
	"wasp/wasp/internal/opcodes"
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
	for it.HasNext() {
		b := it.Byte()
		bytes = append(bytes, b)
		if b == target {
			return bytes, nil
		}
	}
	return nil, fmt.Errorf("target byte 0x%x not found", target)
}

// All integer constants are encoded using a space-efficient, variable-length LEB128 encoding
// unrolled by hand for performance
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

	b = it.data[it.pos+4]
	x |= uint64(b&0x7F) << 28
	if b < 0x80 {
		it.pos += 5
		return int(x)
	}

	b = it.data[it.pos+5]
	x |= uint64(b&0x7F) << 35
	if b < 0x80 {
		it.pos += 6
		return int(x)
	}

	b = it.data[it.pos+6]
	x |= uint64(b&0x7F) << 42
	if b < 0x80 {
		it.pos += 7
		return int(x)
	}

	b = it.data[it.pos+7]
	x |= uint64(b&0x7F) << 49
	if b < 0x80 {
		it.pos += 8
		return int(x)
	}

	b = it.data[it.pos+8]
	x |= uint64(b&0x7F) << 56
	if b < 0x80 {
		it.pos += 9
		return int(x)
	}

	b = it.data[it.pos+9]
	x |= uint64(b&0x7F) << 63
	if b < 0x80 {
		it.pos += 10
		return int(x)
	}

	panic("invalid varint")
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

func (it *Iterator) Position() int {
	return it.pos
}

func (it *Iterator) Seek(pos int) {
	it.pos = pos
}

func (it *Iterator) HasNext() bool {
	return it.pos < it.size
}

func (it *Iterator) Float32() float32 {
	bits := it.Uint32()
	return castPointer[float32](bits)
}

func (it *Iterator) Float64() float64 {
	bits := it.Uint64()
	return castPointer[float64](bits)
}

func (it *Iterator) Uint64() uint64 {
	value := binary.LittleEndian.Uint64(it.data[it.pos:])
	it.pos += 8
	return value
}

func (it *Iterator) Range(start, end int) []byte {
	return it.data[start:end]
}

func (it *Iterator) Opcode() opcodes.Opcode {
	b := uint16(it.Byte())
	if b == 0xFC || b == 0xFD {
		b2 := uint16(it.Byte())
		return opcodes.Opcode(b<<8 | b2)
	}
	return opcodes.Opcode(b)
}

func (it *Iterator) Move(n int) {
	it.pos += n
}

func castPointer[T, S any](bits S) T {
	return *(*T)(unsafe.Pointer(&bits))
}

func NewIterator(data []byte) *Iterator {
	return &Iterator{
		data: data,
		size: len(data),
		pos:  0,
	}
}
