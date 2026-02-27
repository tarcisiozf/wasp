package bar

import (
	"iter"
)

type Segment struct {
	offset, end int
	size        int
	parent      *Segment
	left, right *Segment
	data        []byte
}

func (s Segment) set(offset int, data []byte) {
	start := offset - s.offset
	copy(s.data[start:], data)
}

type Foo struct {
	root           *Segment
	chunkThreshold int
}

func NewFoo(chunkThreshold int) *Foo {
	return &Foo{
		chunkThreshold: chunkThreshold,
	}
}

func (foo *Foo) Store(offset int, data []byte) {
	for start, end := range foo.chunkify(data) {
		chunkOffset := offset + start
		//chunkSize := end - start

		segments := foo.segmentsForRange(chunkOffset)

		if len(segments) == 0 {
			foo.insertSegment(chunkOffset, data[start:end])
			continue
		}

		seg := foo.mergeSegments(segments)
		seg.set(chunkOffset, data[start:end])
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

func (foo *Foo) segmentsForRange(offset int) (segments []*Segment) {
	if foo.root == nil {
		return nil
	}

	current := foo.root
	for current != nil {
		if offset < current.offset {
			current = current.left
		} else if offset >= current.end {
			current = current.right
		} else {
			segments = append(segments, current)
			current = current.right
		}
	}

	return segments
}

func (foo *Foo) insertSegment(offset int, data []byte) {
	size := len(data)
	seg := &Segment{
		offset: offset,
		end:    offset + size,
		size:   size,
		data:   make([]byte, size),
	}
	seg.set(offset, data)

	if foo.root == nil {
		foo.root = seg
		return
	}

	current := foo.root
	for current != nil {
		if seg.offset < current.offset {
			if current.left == nil {
				current.left = seg
				seg.parent = current
				return
			}
			current = current.left
		} else {
			if current.right == nil {
				current.right = seg
				seg.parent = current
				return
			}
			current = current.right
		}
	}
}

func (foo *Foo) mergeSegments(segments []*Segment) *Segment {
	panic("not implemented")
}
