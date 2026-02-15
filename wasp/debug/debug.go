package debug

import (
	"fmt"
	"math"
	"strings"
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/opcodes"
)

func FuncToString(data []byte, index, offset int) (str string, err error) {
	iter := binary.NewIterator(data)
	str = fmt.Sprintf("; function body %d\n", index)
	str += f(offset, "func body size", len(data))

	localDeclCount := iter.Varint()
	str += f(offset+iter.Position(), "local decl count", localDeclCount)

	for i := 0; i < localDeclCount; i++ {
		localTypeCount := iter.Varint()
		str += f(offset+iter.Position(), "local type count", localTypeCount)

		localType := iter.Byte()
		str += f(offset+iter.Position(), typeToString(localType), localType)
	}

	for iter.HasNext() {
		opcode := iter.Opcode()
		str += f(offset+iter.Position(), opcodeName(opcode), opcode)
		lines, err := k(iter, offset, opcode)
		if err != nil {
			return str, err
		}
		str += lines
	}
	return str, nil
}

func opcodeName(opcode opcodes.Opcode) string {
	return opcodes.Name(opcode)
}

func k(iter *binary.Iterator, offset int, opcode opcodes.Opcode) (str string, err error) {
	switch opcode {
	case opcodes.GlobalGet, opcodes.GlobalSet:
		return f(offset+iter.Position(), "global index", iter.Varint()), nil

	case opcodes.I32Const:
		return f(offset+iter.Position(), "i32 literal", iter.Varint()), nil

	case opcodes.I64Const:
		return f(offset+iter.Position(), "i64 literal", iter.Varint()), nil

	case opcodes.F64Const:
		return f(offset+iter.Position(), "f64 literal", iter.Float64()), nil

	case opcodes.F32Const:
		return f(offset+iter.Position(), "f32 literal", iter.Float32()), nil

	case opcodes.LocalTee, opcodes.LocalGet, opcodes.LocalSet:
		return f(offset+iter.Position(), "local index", iter.Varint()), nil

	case opcodes.Block, opcodes.Loop:
		return f(offset+iter.Position(), typeToString(iter.Peek()), iter.Byte()), nil

	case opcodes.Br, opcodes.BrIf:
		return f(offset+iter.Position(), "break depth", iter.Varint()), nil

	case opcodes.I32Load8U, opcodes.I32Load8S, opcodes.I32Load,
		opcodes.I64Load, opcodes.I32Load16S, opcodes.I32Load16U,
		opcodes.I64Load8S, opcodes.I64Load8U, opcodes.I64Load16S,
		opcodes.I64Load16U, opcodes.I64Load32S, opcodes.I64Load32U,
		opcodes.F64Load, opcodes.F32Load:
		return f(offset+iter.Position(), "alignment", iter.Byte()) +
			f(offset+iter.Position(), "load offset", iter.Varint()), nil

	case opcodes.I32Store, opcodes.I64Store, opcodes.I32Store16,
		opcodes.I32Store8, opcodes.I64Store16, opcodes.I64Store8,
		opcodes.I64Store32, opcodes.F64Store, opcodes.F32Store:
		return f(offset+iter.Position(), "alignment", iter.Byte()) +
			f(offset+iter.Position(), "store offset", iter.Varint()), nil

	case opcodes.Call, opcodes.ReturnCall:
		return f(offset+iter.Position(), "function index", iter.Varint()), nil

	case opcodes.CallIndirect:
		return f(offset+iter.Position(), "signature index", iter.Varint()) +
			f(offset+iter.Position(), "table index", iter.Varint()), nil

	case opcodes.BrTable:
		numTargets := iter.Varint()
		str = f(offset+iter.Position(), "num targets", numTargets)
		for i := 0; i < numTargets; i++ {
			str += f(offset+iter.Position(), "break depth", iter.Varint())
		}
		return str + f(offset+iter.Position(), "break depth for default", iter.Varint()), nil

	case opcodes.MemoryFill, opcodes.MemorySize, opcodes.MemoryGrow:
		return f(offset+iter.Position(), "memidx", iter.Varint()), nil

	case opcodes.MemoryCopy:
		return f(offset+iter.Position(), "dst memidx", iter.Varint()) +
			f(offset+iter.Position(), "src memidx", iter.Varint()), nil

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
		opcodes.Unreachable, opcodes.Return:
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
	default:
		panic(fmt.Sprintf("unknown type: %02x", b))
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
	default:
		panic(fmt.Sprintf("unknown type %T", x))
	}
}

func leb128encode(value int) string {
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
