package stack

import (
	"fmt"
	"math"
)

type Types interface {
	int | uint64 | int8 | int16 | int32 | int64 | float32 | float64
}

type Kind uint8

const (
	kindInt Kind = iota
	kindInt64
	kindUint64
	kindInt32
	kindFloat32
	kindFloat64
)

type Entry struct {
	kind  Kind
	value uint64
}

func (e *Entry) I32() int32  { return int32(e.value) }
func (e *Entry) U32() uint32 { return uint32(e.value) }

type Stack struct {
	entries []Entry
}

func New() *Stack {
	return NewWithCapacity(0)
}

func NewWithCapacity(cap int) *Stack {
	return &Stack{
		entries: make([]Entry, 0, cap),
	}
}

func (s *Stack) Push(value any) {
	s.push(value)
}

func (s *Stack) PushInt32(value int32) {
	s.pushEntry(kindInt32, uint64(value))
}

func (s *Stack) PushMany(values ...any) {
	for _, v := range values {
		s.push(v)
	}
}

func (s *Stack) push(value any) {
	var k Kind
	var v uint64

	switch value := value.(type) {
	case int:
		v = uint64(value)
		k = kindInt
	case int64:
		v = uint64(value)
		k = kindInt64
	case uint64:
		v = value
		k = kindUint64
	case int32:
		v = uint64(value)
		k = kindInt32
	case float32:
		v = uint64(math.Float32bits(value))
		k = kindFloat32
	case float64:
		v = math.Float64bits(value)
		k = kindFloat64
	default:
		panic(fmt.Sprintf("unsupported type: %T", value))
	}

	s.pushEntry(k, v)
}

func (s *Stack) pushEntry(kind Kind, value uint64) {
	s.entries = append(s.entries, Entry{kind: kind, value: value})
}

func (s *Stack) Size() int {
	return len(s.entries)
}

func (s *Stack) Last(n int) []any {
	if n > len(s.entries) {
		panic("index out of range")
	}
	out := make([]any, n)
	for i := 0; i < n; i++ {
		out[i] = s.any(len(s.entries) - n + i)
	}
	s.entries = s.entries[:len(s.entries)-n]
	return out
}

func (s *Stack) PeekLast(n int) []any {
	if n > len(s.entries) {
		panic("index out of range")
	}
	out := make([]any, n)
	for i := 0; i < n; i++ {
		out[i] = s.any(len(s.entries) - n + i)
	}
	return out
}

func (s *Stack) Pop() any {
	if len(s.entries) == 0 {
		panic("stack underflow")
	}
	idx := len(s.entries) - 1
	value := s.any(idx)
	s.entries = s.entries[:idx]
	return value
}

func (s *Stack) Peek() any {
	if len(s.entries) == 0 {
		panic("stack underflow")
	}
	return s.any(len(s.entries) - 1)
}

func (s *Stack) any(idx int) any {
	e := s.entries[idx]

	switch e.kind {
	case kindInt:
		return int(e.value)
	case kindInt64:
		return int64(e.value)
	case kindUint64:
		return e.value
	case kindInt32:
		return int32(e.value)
	case kindFloat32:
		return math.Float32frombits(uint32(e.value))
	case kindFloat64:
		return math.Float64frombits(e.value)
	default:
		panic("unsupported type")
	}
}

func (s *Stack) Drop() {
	if len(s.entries) == 0 {
		panic("stack underflow")
	}
	s.entries = s.entries[:len(s.entries)-1]
}

func (s *Stack) At(idx int) any {
	if idx < 0 || idx >= len(s.entries) {
		panic("index out of range")
	}
	return s.any(idx)
}

func (s *Stack) Top() any {
	if len(s.entries) == 0 {
		return 0
	}
	return s.any(len(s.entries) - 1)
}

func (s *Stack) Set(index int, entry Entry) {
	s.entries[index] = entry
}

func (s *Stack) PopEntry() Entry {
	idx := len(s.entries) - 1
	entry := s.entries[idx]
	s.entries = s.entries[:idx]
	return entry
}

func (s *Stack) PeekEntry() Entry {
	return s.entries[len(s.entries)-1]
}

func (s *Stack) AtEntry(index int) Entry {
	return s.entries[index]
}

func (s *Stack) PushEntry(value Entry) {
	s.pushEntry(value.kind, value.value)
}

func Pop[T Types](s *Stack) T {
	if len(s.entries) == 0 {
		panic("stack underflow")
	}

	last := len(s.entries) - 1
	e := s.entries[last]
	s.entries = s.entries[:last]

	switch e.kind {
	case kindInt:
		return T(int(e.value))
	case kindInt64:
		return T(int64(e.value))
	case kindUint64:
		return T(e.value)
	case kindInt32:
		return T(int32(e.value))
	case kindFloat32:
		return T(math.Float32frombits(uint32(e.value)))
	case kindFloat64:
		return T(math.Float64frombits(e.value))
	default:
		panic("unsupported type")
	}
}
