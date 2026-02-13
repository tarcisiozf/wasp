package lex

import (
	"fmt"
	"wasp/cmd/wapo/parser/tokens"
)

type Iterator struct {
	data []byte
	size int
	pos  int
}

func (it *Iterator) HasNext() bool {
	return it.pos < it.size
}

func (it *Iterator) Peek() byte {
	return it.data[it.pos]
}

func (it *Iterator) SkipLine() {
	for it.HasNext() {
		if it.Next() == tokens.Newline {
			break
		}
	}
}

func (it *Iterator) Next() byte {
	b := it.data[it.pos]
	it.pos++
	return b
}

func (it *Iterator) Range(start, end int) []byte {
	return it.data[start:end]
}

func (it *Iterator) Position() int {
	return it.pos
}

func (it *Iterator) Line(index int) int {
	var line int
	for i := 0; i < index; i++ {
		if it.data[i] == tokens.Newline {
			line++
		}
	}
	return line + 1
}

func (it *Iterator) Col(index int) int {
	var line int
	for i := 0; i < index; i++ {
		if it.data[i] == tokens.Newline {
			line = i
		}
	}
	return index - line
}

func (it *Iterator) PeekAt(delta int) byte {
	return it.data[it.pos+delta]
}

func (it *Iterator) String() string {
	str := fmt.Sprintf("line: %d, col: %d\n", it.Line(it.pos), it.Col(it.pos))
	start := it.pos - 20
	if start < 0 {
		start = 0
	}
	end := it.pos + 10
	if end > it.size {
		end = it.size
	}
	line := make([]byte, 0, end-start)
	for i := start; i < end; i++ {
		if it.data[i] == tokens.Newline {
			line = append(line, '\\', 'n')
		} else {
			line = append(line, it.data[i])
		}
	}
	str += string(line) + "\n"
	str += fmt.Sprintf("%s^", string(make([]byte, it.pos-start+1)))
	str += "\n-----------------\n"
	return str
}

func (it *Iterator) PeekPrev() byte {
	return it.data[it.pos-1]
}

func (it *Iterator) PeekNext() byte {
	return it.data[it.pos+1]
}

func (it *Iterator) ReadUntil(b byte) {
	for it.HasNext() {
		if it.Next() == b {
			break
		}
	}
}

func NewIterator(data []byte) *Iterator {
	return &Iterator{
		data: data,
		size: len(data),
		pos:  0,
	}
}
