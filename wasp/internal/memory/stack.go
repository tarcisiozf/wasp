package memory

type Stack[T any] struct {
	items []T
}

func NewStack[T any](items ...T) *Stack[T] {
	return &Stack[T]{
		items: items,
	}
}

func NewStackWithCapacity[T any](capacity int) *Stack[T] {
	return &Stack[T]{
		items: make([]T, 0, capacity),
	}
}

func (s *Stack[T]) Push(value ...T) {
	s.items = append(s.items, value...)
}

func (s *Stack[T]) Pop() T {
	size := len(s.items)
	if size == 0 {
		panic("stack underflow")
	}

	value := s.items[size-1]
	s.items = s.items[:size-1]
	return value
}

func (s *Stack[T]) Last(n int) []T {
	size := len(s.items)
	if n > size {
		panic("stack underflow")
	}
	items := s.items[size-n:]
	s.items = s.items[:size-n]
	return items
}

func (s *Stack[T]) Peek() T {
	pos := len(s.items) - 1
	if pos < 0 {
		panic("stack underflow")
	}
	return s.items[pos]
}

func (s *Stack[T]) At(index int) T {
	return s.items[index]
}

func (s *Stack[T]) Set(index int, item T) {
	s.items[index] = item
}

func (s *Stack[T]) Size() int {
	return len(s.items)
}

func (s *Stack[T]) Items() []T {
	return s.items[:]
}

func (s *Stack[T]) Top() T {
	if len(s.items) == 0 {
		var zero T
		return zero
	}
	return s.items[len(s.items)-1]
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

func (s *Stack[T]) PopN(n int) []T {
	size := len(s.items)
	if n > size {
		panic("stack underflow")
	}
	items := make([]T, n)
	for i := range items {
		items[i] = s.items[size-n+i]
	}
	s.items = s.items[:size-n]
	return items
}
