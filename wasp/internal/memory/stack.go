package memory

type Stack struct {
	items []any
}

func NewStack() *Stack {
	return &Stack{}
}

func (s *Stack) Push(value any) {
	s.items = append(s.items, value)
}

func (s *Stack) Pop() any {
	size := len(s.items)
	if size == 0 {
		panic("stack underflow")
	}

	value := s.items[size-1]
	s.items = s.items[:size-1]
	return value
}

func (s *Stack) PopN(n int) []any {
	size := len(s.items)
	if n > size {
		panic("stack underflow")
	}
	items := s.items[size-n:]
	s.items = s.items[:size-n]
	return items
}
