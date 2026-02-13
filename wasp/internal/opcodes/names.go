package opcodes

import "fmt"

var names = map[byte]string{
	Nop:   "nop",
	Block: "block",
	If:    "if",
	Else:  "else",
	Call:  "call",
	End:   "end",
	Br:    "br",

	LocalGet: "local.get",
	LocalSet: "local.set",
	LocalTee: "local.tee",

	GlobalGet: "global.get",
	GlobalSet: "global.set",

	Const: "const",

	EqI32:  "i32.eq",
	I32Add: "i32.add",
	I32Mul: "i32.mul",
	I32And: "i32.and",
}

func Name(opcode byte) string {
	name, ok := names[opcode]
	if !ok {
		return fmt.Sprintf("unknow(0x%x)", opcode)
	}
	return name
}
