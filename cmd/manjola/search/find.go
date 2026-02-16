package search

import (
	"iter"
	"wasp/cmd/manjola/ast"
)

func FindByType[T ast.Node](nodes iter.Seq[ast.Node]) (T, bool) {
	for node := range nodes {
		switch node.(type) {
		case T:
			return node.(T), true
		}
	}
	var zero T
	return zero, false
}
