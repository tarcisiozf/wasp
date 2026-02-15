package debug

import (
	"fmt"
	"math"
	"strings"
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/opcodes"
)

var sections = map[byte]string{
	0x0: "Custom",
	0x1: "Type",
	0x2: "Import",
	0x3: "Function",
	0x4: "Table",
	0x5: "Memory",
	0x6: "Global",
	0x7: "Export",
	0x8: "Start",
	0x9: "Element",
	0xa: "Code",
	0xb: "Data",
	0xc: "DataCount",
	0xd: "Tag",
}

func WasmToString(data []byte) (str string, err error) {
	iter := binary.NewIterator(data)
	str = f(iter.Position(), "WASM_BINARY_MAGIC", iter.Bytes(4))
	str += f(iter.Position(), "WASM_BINARY_VERSION", iter.Bytes(4))

	for iter.HasNext() {
		pos := iter.Position()
		sectionID := iter.Byte()
		str += fmt.Sprintf("; section \"%s\" (%d)\n", sections[sectionID], sectionID)
		str += f(pos, "section code", sectionID)

		pos = iter.Position()
		sectionSize := iter.Varint()
		str += f(pos, "section size", sectionSize)

		if sectionSize == 0 {
			panic("section size is zero")
		}

		var section string
		switch sectionID {
		case 0x0: // Custom section
			section = sectionCustomToString(iter, sectionSize)
		case 0x1: // Type section
			section = sectionTypeToString(iter)
		case 0x3: // Function section
			section = sectionFunctionToString(iter)
		case 0x7: // Export section
			section = sectionExportToString(iter)
		case 0xa: // Code section
			section, err = sectionCodeToString(iter)
		default:
			return str, fmt.Errorf("invalid section type: %d", sectionID)
		}

		if err != nil {
			return str, fmt.Errorf("failed to parse section %s: %v", sections[sectionID], err)
		}

		str += section
	}

	return str, nil
}

func sectionCustomToString(iter *binary.Iterator, sectionSize int) string {
	startPos := iter.Position()

	nameLen := iter.Varint()
	str := f(startPos, "custom section name length", nameLen)

	pos := iter.Position()
	name := iter.Bytes(nameLen)
	str += f(pos, fmt.Sprintf("custom section name: %s", name), name)

	// Calculate remaining bytes in the section
	bytesRead := iter.Position() - startPos
	dataLen := sectionSize - bytesRead
	data := iter.Bytes(dataLen)
	str += f(iter.Position(), fmt.Sprintf("custom section data (%d bytes)", dataLen), data)

	return str
}

func sectionCodeToString(iter *binary.Iterator) (string, error) {
	pos := iter.Position()
	numFunctions := iter.Varint()
	str := f(pos, "num functions", numFunctions)

	for i := 0; i < numFunctions; i++ {
		fn, err := funcToString(iter, i)
		if err != nil {
			return str, fmt.Errorf("failed to parse function %d: %v", i, err)
		}
		str += fn
	}

	return str, nil
}

func sectionExportToString(iter *binary.Iterator) string {
	pos := iter.Position()
	numExports := iter.Varint()
	str := f(pos, "num exports", numExports)

	for i := 0; i < numExports; i++ {
		pos = iter.Position()
		nameLen := iter.Varint()
		str += f(pos, "string length", nameLen)

		pos = iter.Position()
		name := iter.Bytes(nameLen)
		str += f(pos, fmt.Sprintf("export name %s", name), name)

		pos = iter.Position()
		kind := iter.Byte()
		str += f(pos, "export kind", kind)

		pos = iter.Position()
		index := iter.Varint()
		str += f(pos, fmt.Sprintf("export %s index", kindToString(kind)), index)
	}

	return str
}

func sectionFunctionToString(iter *binary.Iterator) string {
	pos := iter.Position()
	numFunctions := iter.Varint()
	str := f(pos, "num functions", numFunctions)
	for i := 0; i < numFunctions; i++ {
		pos = iter.Position()
		typeIndex := iter.Varint()
		str += f(pos, fmt.Sprintf("function %d signature index", i), typeIndex)
	}
	return str
}

func sectionTypeToString(iter *binary.Iterator) string {
	numTypes := iter.Varint()
	str := f(iter.Position(), "num types", numTypes)
	for i := 0; i < numTypes; i++ {
		pos := iter.Position()
		form := iter.Byte()

		str += fmt.Sprintf("; %s type %d\n", typeToString(form), i)
		str += f(pos, typeToString(form), form)

		switch form {
		case 0x60: // func type
			pos = iter.Position()
			numParams := iter.Varint()

			str += f(pos, "num params", numParams)
			for j := 0; j < numParams; j++ {
				pos = iter.Position()
				paramType := iter.Byte()
				str += f(pos, typeToString(paramType), paramType)
			}

			pos = iter.Position()
			numResults := iter.Varint()
			str += f(pos, "num results", numResults)
			for j := 0; j < numResults; j++ {
				pos = iter.Position()
				resultType := iter.Byte()
				str += f(pos, typeToString(resultType), resultType)
			}
		default:
			panic(fmt.Sprintf("unknown type form: %02x", form))
		}
	}
	return str
}

func funcToString(iter *binary.Iterator, index int) (str string, err error) {
	str = fmt.Sprintf("; function body %d\n", index)

	pos := iter.Position()
	bodySize := iter.Varint()
	str += f(pos, "func body size", bodySize)

	if bodySize == 0 {
		return str, fmt.Errorf("invalid body size: %d", bodySize)
	}

	pos = iter.Position()
	localDeclCount := iter.Varint()
	str += f(pos, "local decl count", localDeclCount)

	for i := 0; i < localDeclCount; i++ {
		str += f(iter.Position(), "local type count", iter.Varint())

		pos = iter.Position()
		localType := iter.Byte()
		str += f(pos, typeToString(localType), localType)
	}

	var depth int
	for iter.HasNext() {
		opcode := iter.Opcode()
		str += f(iter.Position(), opcodeName(opcode), opcode)

		lines, err := k(iter, opcode)
		if err != nil {
			return str, err
		}
		str += lines

		if isBranchingOpcode(opcode) {
			depth++
		} else if opcode == opcodes.End {
			if depth == 0 {
				break
			}
			depth--
		}
	}
	return str, nil
}

func isBranchingOpcode(opcode opcodes.Opcode) bool {
	return opcode == opcodes.If || opcode == opcodes.Block || opcode == opcodes.Loop || opcode == opcodes.Br || opcode == opcodes.BrIf || opcode == opcodes.BrTable
}

func opcodeName(opcode opcodes.Opcode) string {
	return opcodes.Name(opcode)
}

func k(iter *binary.Iterator, opcode opcodes.Opcode) (str string, err error) {
	switch opcode {
	case opcodes.GlobalGet, opcodes.GlobalSet:
		return f(iter.Position(), "global index", iter.Varint()), nil

	case opcodes.I32Const:
		return f(iter.Position(), "i32 literal", iter.Varint()), nil

	case opcodes.I64Const:
		return f(iter.Position(), "i64 literal", iter.Varint()), nil

	case opcodes.F64Const:
		return f(iter.Position(), "f64 literal", iter.Float64()), nil

	case opcodes.F32Const:
		return f(iter.Position(), "f32 literal", iter.Float32()), nil

	case opcodes.LocalTee, opcodes.LocalGet, opcodes.LocalSet:
		return f(iter.Position(), "local index", iter.Varint()), nil

	case opcodes.Block, opcodes.Loop:
		return f(iter.Position(), typeToString(iter.Peek()), iter.Byte()), nil

	case opcodes.Br, opcodes.BrIf:
		return f(iter.Position(), "break depth", iter.Varint()), nil

	case opcodes.I32Load8U, opcodes.I32Load8S, opcodes.I32Load,
		opcodes.I64Load, opcodes.I32Load16S, opcodes.I32Load16U,
		opcodes.I64Load8S, opcodes.I64Load8U, opcodes.I64Load16S,
		opcodes.I64Load16U, opcodes.I64Load32S, opcodes.I64Load32U,
		opcodes.F64Load, opcodes.F32Load:
		return f(iter.Position(), "alignment", iter.Byte()) +
			f(iter.Position(), "load offset", iter.Varint()), nil

	case opcodes.I32Store, opcodes.I64Store, opcodes.I32Store16,
		opcodes.I32Store8, opcodes.I64Store16, opcodes.I64Store8,
		opcodes.I64Store32, opcodes.F64Store, opcodes.F32Store:
		return f(iter.Position(), "alignment", iter.Byte()) +
			f(iter.Position(), "store offset", iter.Varint()), nil

	case opcodes.Call, opcodes.ReturnCall:
		return f(iter.Position(), "function index", iter.Varint()), nil

	case opcodes.CallIndirect:
		return f(iter.Position(), "signature index", iter.Varint()) +
			f(iter.Position(), "table index", iter.Varint()), nil

	case opcodes.BrTable:
		numTargets := iter.Varint()
		str = f(iter.Position(), "num targets", numTargets)
		for i := 0; i < numTargets; i++ {
			str += f(iter.Position(), "break depth", iter.Varint())
		}
		return str + f(iter.Position(), "break depth for default", iter.Varint()), nil

	case opcodes.MemoryFill, opcodes.MemorySize, opcodes.MemoryGrow:
		return f(iter.Position(), "memidx", iter.Varint()), nil

	case opcodes.MemoryCopy:
		return f(iter.Position(), "dst memidx", iter.Varint()) +
			f(iter.Position(), "src memidx", iter.Varint()), nil

	case opcodes.If:
		return f(iter.Position(), typeToString(iter.Peek()), iter.Byte()), nil

	case opcodes.I32Sub, opcodes.I32Add, opcodes.I32Or, opcodes.I32Xor,
		opcodes.I32GtS, opcodes.I32Eqz, opcodes.I32And, opcodes.I32Ne,
		opcodes.I32Eq, opcodes.I32GtU, opcodes.I32LeU, opcodes.I32Shl,
		opcodes.I32GeU, opcodes.I32Mul, opcodes.I32LtU, opcodes.I32LtS,
		opcodes.I32Extend8S, opcodes.I32WrapI64, opcodes.I32ShrU,
		opcodes.I32DivS, opcodes.I32DivU, opcodes.I32ShrS, opcodes.I32RemS,
		opcodes.I32RemU, opcodes.I32Clz, opcodes.I32LeS, opcodes.I32GeS,
		opcodes.I32Extend16S, opcodes.I32ReinterpretF32,
		opcodes.I32TruncSatF32S, opcodes.I32Rotr, opcodes.I32Rotl:
		// no immediate arguments
		return

	case opcodes.I64ExtendI32S, opcodes.I64Eqz, opcodes.I64GtU,
		opcodes.I64ShrU, opcodes.I64GtS, opcodes.I64Sub, opcodes.I64Add,
		opcodes.I64And, opcodes.I64Ne, opcodes.I64Eq, opcodes.I64LeU,
		opcodes.I64Shl, opcodes.I64GeU, opcodes.I64Mul, opcodes.I64LtU,
		opcodes.I64LtS, opcodes.I64ExtendI32U, opcodes.I64ShrS,
		opcodes.I64DivU, opcodes.I64DivS, opcodes.I64Xor, opcodes.I64Or,
		opcodes.I64Ctz, opcodes.I64Clz, opcodes.I64RemS, opcodes.I64RemU,
		opcodes.I64Rotl, opcodes.I64Rotr, opcodes.I64GeS, opcodes.I64LeS,
		opcodes.I64ReinterpretF64, opcodes.I64Extend8S,
		opcodes.I64Extend16S, opcodes.I64Extend32S:
		// no immediate arguments
		return

	case opcodes.F64Mul, opcodes.F64Add, opcodes.F64Sub,
		opcodes.F64Div, opcodes.F64Min, opcodes.F64Max,
		opcodes.F64Copysign, opcodes.F64Sqrt, opcodes.F64Ceil,
		opcodes.F64Floor, opcodes.F64Trunc, opcodes.F64Nearest,
		opcodes.F64ReinterpretI64, opcodes.F64PromoteF32,
		opcodes.F64ConvertI32S, opcodes.F64ConvertI32U,
		opcodes.F64ConvertI64S, opcodes.F64ConvertI64U,
		opcodes.F64Abs, opcodes.F64Lt, opcodes.F64Gt,
		opcodes.F64Le, opcodes.F64Ge, opcodes.F64Eq,
		opcodes.F64Ne:
		// no immediate arguments
		return

	case opcodes.F32ConvertI32S, opcodes.F32ConvertI32U, opcodes.F32ConvertI64S,
		opcodes.F32ConvertI64U, opcodes.F32DemoteF64, opcodes.F32Mul,
		opcodes.F32Add, opcodes.F32Sub, opcodes.F32Div, opcodes.F32Min,
		opcodes.F32Max, opcodes.F32Copysign, opcodes.F32Sqrt,
		opcodes.F32Ceil, opcodes.F32Floor, opcodes.F32Trunc,
		opcodes.F32Nearest, opcodes.F32ReinterpretI32,
		opcodes.F32Abs, opcodes.F32Lt, opcodes.F32Gt, opcodes.F32Le,
		opcodes.F32Ge, opcodes.F32Eq, opcodes.F32Ne:
		// no immediate arguments
		return

	case opcodes.End, opcodes.Drop, opcodes.Select,
		opcodes.Unreachable, opcodes.Return, opcodes.Else:
		// no immediate arguments
		return

	default:
		return "", fmt.Errorf("unknown opcode: %02x\n", opcode)
	}
}

func typeToString(b byte) string {
	switch b {
	case 0x7F:
		return "i32"
	case 0x7E:
		return "i64"
	case 0x7D:
		return "f32"
	case 0x7C:
		return "f64"
	case 0x40:
		return "void"
	case 0x60:
		return "func"
	default:
		panic(fmt.Sprintf("unknown type: %02x", b))
	}
}

func kindToString(b byte) string {
	switch b {
	case 0x00:
		return "function"
	case 0x01:
		return "table"
	case 0x02:
		return "memory"
	case 0x03:
		return "global"
	default:
		panic(fmt.Sprintf("unknown export kind: %02x", b))
	}
}

func f(offset int, label string, x any) string {
	str := fmt.Sprintf("%s: %s", p(offset), valueToHex(x))
	return fmt.Sprintf("%s; %s\n", pad(str, 50), label)
}

func pad(str string, n int) string {
	size := len(str)
	if size >= n {
		return str
	}
	return str + strings.Repeat(" ", n-size)
}

func p(offset int) string {
	return fmt.Sprintf("%08x", offset)
}

func valueToHex(x any) string {
	switch x.(type) {
	case byte:
		return fmt.Sprintf("%02x", x.(byte))
	case uint16:
		if x.(uint16) <= 0xFF {
			return fmt.Sprintf("%02x", x.(uint16))
		}
		return fmt.Sprintf("%04x", x.(uint16))
	case int:
		return leb128encode(x.(int))
	case float32:
		return fmt.Sprintf("%08x", math.Float32bits(x.(float32)))
	case float64:
		return fmt.Sprintf("%016x", math.Float64bits(x.(float64)))
	case []byte:
		return fmt.Sprintf("%x", x.([]byte))
	default:
		panic(fmt.Sprintf("unknown type %T", x))
	}
}

func leb128encode(value int) string {
	if value == 0 {
		return "00"
	}
	bytes := make([]byte, 10)
	var off int
	var v = uint64(value)
	for v != 0 {
		b := byte(v & 0x7F)
		v >>= 7
		if v != 0 {
			b |= 0x80
		}
		bytes[off] = b
		off++
	}
	return fmt.Sprintf("%x", bytes[:off])
}
