package wasp

type Stack struct {
	items []any
}

func newStack() *Stack {
	return &Stack{}
}

func (s *Stack) push(value any) {
	s.items = append(s.items, value)
}

func (s *Stack) pop() any {
	if len(s.items) == 0 {
		panic("stack underflow")
	}

	value := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return value
}
