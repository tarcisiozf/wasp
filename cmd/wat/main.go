package main

import (
	"fmt"
	"os"
	"wasp/cmd/wat/wat/parser"
)

func main() {
	//if len(os.Args) < 2 {
	//	println("Usage: wat <file.wat>")
	//	os.Exit(1)
	//}
	//
	//filename := os.Args[1]
	//file, err := os.ReadFile(filename)
	//if err != nil {
	//	println("Error reading file:", err.Error())
	//	os.Exit(1)
	//}

	file := []byte(`
;; Local count can be 0.
(module binary
  "\00asm" "\01\00\00\00"
  "\01\04\01\60\00\00"     ;; Type section
  "\03\02\01\00"           ;; Function section
  "\0a\0a\01"              ;; Code section

  ;; function 0
  "\08\03"
  "\00\7f"                 ;; 0 i32
  "\00\7e"                 ;; 0 i64
  "\02\7d"                 ;; 2 f32
  "\0b"                    ;; end
)`)

	root, err := parser.Parse(file)
	if err != nil {
		println("Error parsing file:", err.Error())
		os.Exit(1)
	}
	fmt.Println(root)
}
