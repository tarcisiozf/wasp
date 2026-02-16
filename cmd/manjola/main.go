package main

import (
	"fmt"
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
	rootNodes := ast.Parse(lexer)

	for listNode := range ast.WalkType[*ast.List](rootNodes) {
		if listNode.ElemType != "func" {
			continue
		}
		funcStr := ast.Stringify(listNode)
		fmt.Println(funcStr)
	}
}
