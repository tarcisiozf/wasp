package main

import (
	"fmt"
	"iter"
	"os"
	"wasp/cmd/manjola/ast"
	"wasp/cmd/manjola/lex"
)

func main() {
	if len(os.Args) < 2 {
		println("Usage: wat <file.wat>")
		os.Exit(1)
	}

	filename := os.Args[1]
	file, err := os.ReadFile(filename)
	if err != nil {
		println("Error reading file:", err.Error())
		os.Exit(1)
	}

	lexer := lex.NewLexer(file)
	program := ast.Parse(lexer)
	module, ok := FindByType[*ast.Module](program)
	if !ok {
		println("Error: no module found in the program")
	}
	funcs := FilterChildrenByType[*ast.Func](module)

	for fn := range funcs {
		funcStr := ast.Stringify(fn)
		fmt.Println(funcStr)
	}
}

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
