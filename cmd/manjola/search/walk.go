package search

import (
	"iter"
	"wasp/cmd/manjola/ast"
)

func Walk(seq iter.Seq[ast.Node]) iter.Seq[ast.Node] {
	var queue []ast.Node
	for n := range seq {
		queue = append(queue, n)
	}
	return func(yield func(ast.Node) bool) {
		for len(queue) > 0 {
			node := queue[0]
			queue = queue[1:]

			if !yield(node) {
				return
			}

			children := node.Children()
			if len(children) > 0 {
				queue = append(queue, children...)
			}
		}
	}
}

func WalkType[T ast.Node](seq iter.Seq[ast.Node]) iter.Seq[T] {
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
