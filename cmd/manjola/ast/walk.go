package ast

import (
	"iter"
)

func Walk(seq iter.Seq[Node]) iter.Seq[Node] {
	var queue []Node
	for n := range seq {
		queue = append(queue, n)
	}
	return func(yield func(Node) bool) {
		for len(queue) > 0 {
			node := queue[0]
			queue = queue[1:]

			if !yield(node) {
				return
			}

			switch node.(type) {
			case *List:
				list := node.(*List)
				queue = append(queue, list.Children...)
			}
		}
	}
}

func WalkType[T any](seq iter.Seq[Node]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for node := range Walk(seq) {
			switch node.(type) {
			case T:
				if !yield(node.(T)) {
					return
				}
			}
		}
	}
}
