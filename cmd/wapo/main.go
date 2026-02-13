package main

import (
	"fmt"
	"os"
	"strings"
	"wasp/cmd/wapo/lex"
	"wasp/cmd/wapo/parser"
)

const dir = "/Users/tzf/Desktop/webassembly/spec/test/core/"

func main() {
	files := scanDir(dir, ".wast")
	for _, filename := range files {
		if filename != "annotations.wast" {
			continue
		}
		fmt.Printf("Parsing file: %s\n", filename)
		data, err := os.ReadFile(dir + "/" + filename)
		if err != nil {
			panic(err)
		}

		it := lex.NewIterator(data)
		for node, err := range parser.Parse(it) {
			if err != nil {
				panic(err)
			}
			fmt.Println(node.String())
		}
	}
}

func scanDir(dir, ext string) (files []string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ext) {
			files = append(files, name)
		}
	}
	return files
}
