package main

import (
	"fmt"
	"os"
	"wasp/cmd/manjola/ast"
	"wasp/cmd/manjola/lex"
	"wasp/cmd/manjola/search"
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
	module, ok := search.FindByType[*ast.Module](program)
	if !ok {
		println("Error: no module found in the program")
	}
	funcs := search.FilterChildrenByType[*ast.Func](module)

	for fn := range funcs {
		funcStr := ast.Stringify(fn)
		fmt.Println(funcStr)
	}
}
