package search

import (
	"iter"
	"wasp/cmd/manjola/ast"
)

func FilterChildrenByType[T ast.Node](parent ast.Node) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, child := range parent.Children() {
			switch child.(type) {
			case T:
				if !yield(child.(T)) {
					return
				}
			}
		}
	}
}
