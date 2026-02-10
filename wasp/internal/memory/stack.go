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
	if len(s.items) == 0 {
		panic("stack underflow")
	}

	value := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return value
}

func (s *Stack) PopN(n int) []any {
	if n > len(s.items) {
		panic("stack underflow")
	}
	items := s.items[:n]
	s.items = s.items[n:]
	return items
}
