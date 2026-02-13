package main

import (
	"fmt"
	"iter"
	"os"
	"strings"
)

const dir = "/Users/tzf/Desktop/webassembly/spec/test/core/"

func main() {
	files := scanDir(dir, ".wast")
	for _, filename := range files {
		fmt.Printf("Parsing file: %s\n", filename)
		data, err := os.ReadFile(dir + "/" + filename)
		if err != nil {
			panic(err)
		}

		it := NewIterator(data)
		for node, err := range parse(it) {
			if err != nil {
				panic(err)
			}
			fmt.Println(node.content)
		}
	}
}

func parse(it *Iterator) iter.Seq2[*List, error] {
	return func(yield func(*List, error) bool) {
		for it.HasNext() {
			ch := it.Peek()

			switch ch {
			case semicolon:
				it.SkipLine()
			case openParen:
				yield(parseList(it))
			default:
				it.Next()
			}
		}
	}
}

func parseList(it *Iterator) (*List, error) {
	depth := 0
	inString := false
	start := it.Position()
	end := -1
loop:
	for it.HasNext() {
		switch it.Peek() {
		case quote:
			if it.PeekPrev() != backslash {
				inString = !inString
			}
		case openParen:
			if !inString {
				depth++
			}
		case closeParen:
			depth--
			if depth == 0 {
				end = it.Position() + 1
				break loop
			}
		}
		it.Next()
	}
	if end == -1 {
		return nil, fmt.Errorf("unexpected end of list, line: %d, col: %d", it.Line(start), it.Col(start))
	}
	content := string(it.Range(start, end))
	return NewList(content, start), nil
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
