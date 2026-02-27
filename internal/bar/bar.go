package bar

import (
	"fmt"
	"iter"
)

type Bar struct {
	offset, end int
	size        int
	parent      *Bar
	left, right *Bar
	data        []byte
}

type Foo struct {
	root           *Bar
	chunkThreshold int
}

func NewFoo(chunkThreshold int) *Foo {
	return &Foo{
		chunkThreshold: chunkThreshold,
	}
}

func (foo *Foo) Store(offset, data []byte) {
	for start, end := range foo.chunkify(data) {
		fmt.Println(start, end)
	}
}

func (foo *Foo) chunkify(data []byte) iter.Seq2[int, int] {
	return func(yield func(int, int) bool) {
		size := len(data)
		start := -1

		for i := 0; i < size; i++ {
			if data[i] == 0 {
				if start >= 0 && i-start >= foo.chunkThreshold {
					yield(start, i)
					start = -1
				}
			} else if start < 0 {
				start = i
			}
		}

		if start >= 0 {
			yield(start, size)
		}
	}
}
