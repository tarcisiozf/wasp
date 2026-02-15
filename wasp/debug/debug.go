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

func WasmToString(data []byte) (string, error) {
	iter := binary.NewIterator(data)

	sb := &strings.Builder{}

	f(sb, iter.Position(), "WASM_BINARY_MAGIC", iter.Bytes(4))
	f(sb, iter.Position(), "WASM_BINARY_VERSION", iter.Bytes(4))

	for iter.HasNext() {
		pos := iter.Position()
		sectionID := iter.Byte()
		sb.WriteString(fmt.Sprintf("; section \"%s\" (%d)\n", sections[sectionID], sectionID))
		f(sb, pos, "section code", sectionID)

		pos = iter.Position()
		sectionSize := iter.Varint()
		f(sb, pos, "section size", sectionSize)

		if sectionSize == 0 {
			panic("section size is zero")
		}

		var err error
		switch sectionID {
		case 0x0: // Custom section
			sectionCustomToString(sb, iter, sectionSize)
		case 0x1: // Type section
			sectionTypeToString(sb, iter)
		case 0x2: // Import section
			err = sectionImportToString(sb, iter)
		case 0x3: // Function section
			sectionFunctionToString(sb, iter)
		case 0x4: // Table section
			sectionTableToString(sb, iter)
		case 0x5: // Memory section
			sectionMemoryToString(sb, iter)
		case 0x6: // Global section
			err = sectionGlobalToString(sb, iter)
		case 0x7: // Export section
			sectionExportToString(sb, iter)
		case 0x9: // Element section
			err = sectionElementToString(sb, iter)
		case 0xa: // Code section
			err = sectionCodeToString(sb, iter)
		case 0xb: // Data section
			sectionDataToString(sb, iter)
		default:
			return sb.String(), fmt.Errorf("invalid section type: %x", sectionID)
		}

		if err != nil {
			return sb.String(), fmt.Errorf("failed to parse section %s: %v", sections[sectionID], err)
		}
	}

	return sb.String(), nil
}

func sectionDataToString(sb *strings.Builder, iter *binary.Iterator) {
	pos := iter.Position()
	numDataSegments := iter.Varint()
	f(sb, pos, "num data segments", numDataSegments)

	for i := 0; i < numDataSegments; i++ {
		sb.WriteString(fmt.Sprintf("; data segment header %d\n", i))

		pos = iter.Position()
		flags := iter.Byte()
		f(sb, pos, "segment flags", flags)

		pos = iter.Position()
		opcode := iter.Opcode()
		g(sb, pos, opcode)

		if err := k(sb, iter, opcode); err != nil {
			panic(fmt.Sprintf("failed to parse data segment offset expression: %v", err))
		}

		f(sb, iter.Position(), "end", iter.Opcode())

		pos = iter.Position()
		dataSize := iter.Varint()
		f(sb, pos, "data size", dataSize)

		pos = iter.Position()
		data := iter.Bytes(dataSize)
		f(sb, pos, fmt.Sprintf("data bytes (%d bytes)", dataSize), data)
	}
}

func sectionElementToString(sb *strings.Builder, iter *binary.Iterator) error {
	pos := iter.Position()
	numElemSegments := iter.Varint()
	f(sb, pos, "num element segments", numElemSegments)

	for i := 0; i < numElemSegments; i++ {
		sb.WriteString(fmt.Sprintf("; element segment header %d\n", i))

		pos = iter.Position()
		segmentFlags := iter.Byte()
		f(sb, pos, "segment flags", segmentFlags)

		pos = iter.Position()
		opcode := iter.Opcode()
		g(sb, pos, opcode)

		if err := k(sb, iter, opcode); err != nil {
			return err
		}

		f(sb, iter.Position(), "end", iter.Opcode())

		pos = iter.Position()
		numElements := iter.Varint()
		f(sb, pos, "num elements", numElements)

		for i := 0; i < numElements; i++ {
			f(sb, iter.Position(), "element function index", iter.Varint())
		}
	}

	return nil
}

func sectionGlobalToString(sb *strings.Builder, iter *binary.Iterator) (err error) {
	pos := iter.Position()
	numGlobals := iter.Varint()
	f(sb, pos, "num globals", numGlobals)

	for i := 0; i < numGlobals; i++ {
		sb.WriteString(fmt.Sprintf("; global %d\n", i))

		pos = iter.Position()
		contentType := iter.Byte()
		f(sb, pos, typeToString(contentType), contentType)

		pos = iter.Position()
		mutability := iter.Byte()
		f(sb, pos, "global mutability", mutability)

		switch contentType {
		case 0x7F: // i32
			err = k(sb, iter, iter.Opcode())
		default:
			return fmt.Errorf("unsupported global content type: %02x", contentType)
		}

		if err != nil {
			return fmt.Errorf("failed to parse global initializer: %v", err)
		}
	}

	f(sb, iter.Position(), "global section end opcode", iter.Opcode())

	return nil
}

func sectionMemoryToString(sb *strings.Builder, iter *binary.Iterator) {
	pos := iter.Position()
	numMemories := iter.Varint()
	f(sb, pos, "num memories", numMemories)

	for i := 0; i < numMemories; i++ {
		sb.WriteString(fmt.Sprintf("; memory %d\n", i))

		pos = iter.Position()
		limitsFlags := iter.Byte()
		f(sb, pos, "memory limits flags", limitsFlags)

		pos = iter.Position()
		initial := iter.Varint()
		f(sb, pos, "memory initial size (pages)", initial)

		if limitsFlags&0x01 != 0 {
			pos = iter.Position()
			maximum := iter.Varint()
			f(sb, pos, "memory maximum size (pages)", maximum)
		}
	}
}

func sectionTableToString(sb *strings.Builder, iter *binary.Iterator) {
	pos := iter.Position()
	numTables := iter.Varint()
	f(sb, pos, "num tables", numTables)

	for i := 0; i < numTables; i++ {
		sb.WriteString(fmt.Sprintf("; table %d", i))

		pos = iter.Position()
		elementType := iter.Byte()
		f(sb, pos, kindToString(elementType), elementType)

		pos = iter.Position()
		limitsFlags := iter.Byte()
		f(sb, pos, "table limits flags", limitsFlags)

		pos = iter.Position()
		initial := iter.Varint()
		f(sb, pos, "table initial size", initial)

		if limitsFlags&0x01 != 0 {
			pos = iter.Position()
			maximum := iter.Varint()
			f(sb, pos, "table maximum size", maximum)
		}
	}
}

func sectionImportToString(sb *strings.Builder, iter *binary.Iterator) error {
	pos := iter.Position()
	numImports := iter.Varint()
	f(sb, pos, "num imports", numImports)

	for i := 0; i < numImports; i++ {
		sb.WriteString(fmt.Sprintf("; import header %d\n", i))

		pos = iter.Position()
		moduleLen := iter.Varint()
		f(sb, pos, "string length", moduleLen)

		pos = iter.Position()
		module := iter.Bytes(moduleLen)
		f(sb, pos, fmt.Sprintf("import module name: %s", module), module)

		pos = iter.Position()
		fieldLen := iter.Varint()
		f(sb, pos, "field string length", fieldLen)

		pos = iter.Position()
		field := iter.Bytes(fieldLen)
		f(sb, pos, fmt.Sprintf("import field: %s", field), field)

		pos = iter.Position()
		kind := iter.Byte()
		f(sb, pos, "import kind", kind)

		switch kind {
		case 0x00: // function
			pos = iter.Position()
			typeIndex := iter.Varint()
			f(sb, pos, "import signature index", typeIndex)

		//case 0x01: // table
		//	pos = iter.Position()
		//	elementType := iter.Byte()
		//	f(sb, pos, "import table element type", elementType)
		//
		//	pos = iter.Position()
		//	limitsFlags := iter.Byte()
		//	f(sb, pos, "import table limits flags", limitsFlags)
		//
		//	pos = iter.Position()
		//	initial := iter.Varint()
		//	f(sb, pos, "import table initial size", initial)
		//
		//	if limitsFlags&0x01 != 0 {
		//		pos = iter.Position()
		//		maximum := iter.Varint()
		//		f(sb, pos, "import table maximum size", maximum)
		//	}
		//
		//case 0x02: // memory
		//	pos = iter.Position()
		//	limitsFlags := iter.Byte()
		//	f(sb, pos, "import memory limits flags", limitsFlags)
		//
		//	pos = iter.Position()
		//	initial := iter.Varint()
		//	f(sb, pos, "import memory initial size (pages)", initial)
		//
		//	if limitsFlags&0x01 != 0 {
		//		pos = iter.Position()
		//		maximum := iter.Varint()
		//		f(sb, pos, "import memory maximum size (pages)", maximum)
		//	}
		//
		//case 0x03: // global
		//	pos = iter.Position()
		//	contentType := iter.Byte()
		//	f(sb, pos, "import global content type", contentType)
		//
		//	pos = iter.Position()
		//	mutability := iter.Byte()
		//	f(sb, pos, "import global mutability", mutability)

		default:
			return fmt.Errorf("unknown import kind: %02x", kind)
		}
	}
	return nil
}

func sectionCustomToString(sb *strings.Builder, iter *binary.Iterator, sectionSize int) {
	startPos := iter.Position()

	nameLen := iter.Varint()
	f(sb, startPos, "custom section name length", nameLen)

	pos := iter.Position()
	name := iter.Bytes(nameLen)
	f(sb, pos, fmt.Sprintf("custom section name: %s", name), name)

	// Calculate remaining bytes in the section
	bytesRead := iter.Position() - startPos
	dataLen := sectionSize - bytesRead
	data := iter.Bytes(dataLen)
	f(sb, iter.Position(), fmt.Sprintf("custom section data (%d bytes)", dataLen), data)
}

func sectionCodeToString(sb *strings.Builder, iter *binary.Iterator) error {
	pos := iter.Position()
	numFunctions := iter.Varint()
	f(sb, pos, "num functions", numFunctions)

	for i := 0; i < numFunctions; i++ {
		if err := funcToString(sb, iter, i); err != nil {
			return fmt.Errorf("failed to parse function %d: %v", i, err)
		}
	}

	return nil
}

func sectionExportToString(sb *strings.Builder, iter *binary.Iterator) {
	pos := iter.Position()
	numExports := iter.Varint()
	f(sb, pos, "num exports", numExports)

	for i := 0; i < numExports; i++ {
		pos = iter.Position()
		nameLen := iter.Varint()
		f(sb, pos, "string length", nameLen)

		pos = iter.Position()
		name := iter.Bytes(nameLen)
		f(sb, pos, fmt.Sprintf("export name %s", name), name)

		pos = iter.Position()
		kind := iter.Byte()
		f(sb, pos, "export kind", kind)

		pos = iter.Position()
		index := iter.Varint()
		f(sb, pos, fmt.Sprintf("export %s index", kindToString(kind)), index)
	}
}

func sectionFunctionToString(sb *strings.Builder, iter *binary.Iterator) {
	pos := iter.Position()
	numFunctions := iter.Varint()
	f(sb, pos, "num functions", numFunctions)
	for i := 0; i < numFunctions; i++ {
		pos = iter.Position()
		typeIndex := iter.Varint()
		f(sb, pos, fmt.Sprintf("function %d signature index", i), typeIndex)
	}
}

func sectionTypeToString(sb *strings.Builder, iter *binary.Iterator) {
	numTypes := iter.Varint()
	f(sb, iter.Position(), "num types", numTypes)
	for i := 0; i < numTypes; i++ {
		pos := iter.Position()
		form := iter.Byte()

		sb.WriteString(fmt.Sprintf("; %s type %d\n", typeToString(form), i))
		f(sb, pos, typeToString(form), form)

		switch form {
		case 0x60: // func type
			pos = iter.Position()
			numParams := iter.Varint()

			f(sb, pos, "num params", numParams)
			for j := 0; j < numParams; j++ {
				pos = iter.Position()
				paramType := iter.Byte()
				f(sb, pos, typeToString(paramType), paramType)
			}

			pos = iter.Position()
			numResults := iter.Varint()
			f(sb, pos, "num results", numResults)
			for j := 0; j < numResults; j++ {
				pos = iter.Position()
				resultType := iter.Byte()
				f(sb, pos, typeToString(resultType), resultType)
			}
		default:
			panic(fmt.Sprintf("unknown type form: %02x", form))
		}
	}
}

func funcToString(sb *strings.Builder, iter *binary.Iterator, index int) error {
	sb.WriteString(fmt.Sprintf("; function body %d\n", index))

	pos := iter.Position()
	bodySize := iter.Varint()
	f(sb, pos, "func body size", bodySize)

	if bodySize == 0 {
		return fmt.Errorf("invalid body size: %d", bodySize)
	}

	pos = iter.Position()
	localDeclCount := iter.Varint()
	f(sb, pos, "local decl count", localDeclCount)

	for i := 0; i < localDeclCount; i++ {
		f(sb, iter.Position(), "local type count", iter.Varint())

		pos = iter.Position()
		localType := iter.Byte()
		f(sb, pos, typeToString(localType), localType)
	}

	var depth int
	for iter.HasNext() {
		pos = iter.Position()
		opcode := iter.Opcode()
		g(sb, pos, opcode)

		err := k(sb, iter, opcode)
		if err != nil {
			return err
		}

		if isBranchingOpcode(opcode) {
			depth++
		} else if opcode == opcodes.End {
			if depth == 0 {
				break
			}
			depth--
		}
	}
	return nil
}

func g(sb *strings.Builder, offset int, opcode opcodes.Opcode) {
	f(sb, offset, opcode.String(), opcode)
}

func isBranchingOpcode(opcode opcodes.Opcode) bool {
	return opcode == opcodes.If ||
		opcode == opcodes.Block ||
		opcode == opcodes.Loop
}

func k(sb *strings.Builder, iter *binary.Iterator, opcode opcodes.Opcode) (err error) {
	switch opcode {
	case opcodes.GlobalGet, opcodes.GlobalSet:
		f(sb, iter.Position(), "global index", iter.Varint())

	case opcodes.I32Const:
		f(sb, iter.Position(), "i32 literal", iter.Varint())

	case opcodes.I64Const:
		f(sb, iter.Position(), "i64 literal", iter.Varint())

	case opcodes.F64Const:
		f(sb, iter.Position(), "f64 literal", iter.Float64())

	case opcodes.F32Const:
		f(sb, iter.Position(), "f32 literal", iter.Float32())

	case opcodes.LocalTee, opcodes.LocalGet, opcodes.LocalSet:
		f(sb, iter.Position(), "local index", iter.Varint())

	case opcodes.Block, opcodes.Loop:
		f(sb, iter.Position(), typeToString(iter.Peek()), iter.Byte())

	case opcodes.Br, opcodes.BrIf:
		f(sb, iter.Position(), "break depth", iter.Varint())

	case opcodes.I32Load8U, opcodes.I32Load8S, opcodes.I32Load,
		opcodes.I64Load, opcodes.I32Load16S, opcodes.I32Load16U,
		opcodes.I64Load8S, opcodes.I64Load8U, opcodes.I64Load16S,
		opcodes.I64Load16U, opcodes.I64Load32S, opcodes.I64Load32U,
		opcodes.F64Load, opcodes.F32Load:
		f(sb, iter.Position(), "alignment", iter.Byte())
		f(sb, iter.Position(), "load offset", iter.Varint())

	case opcodes.I32Store, opcodes.I64Store, opcodes.I32Store16,
		opcodes.I32Store8, opcodes.I64Store16, opcodes.I64Store8,
		opcodes.I64Store32, opcodes.F64Store, opcodes.F32Store:
		f(sb, iter.Position(), "alignment", iter.Byte())
		f(sb, iter.Position(), "store offset", iter.Varint())

	case opcodes.Call, opcodes.ReturnCall:
		f(sb, iter.Position(), "function index", iter.Varint())

	case opcodes.CallIndirect:
		f(sb, iter.Position(), "signature index", iter.Varint())
		f(sb, iter.Position(), "table index", iter.Varint())

	case opcodes.BrTable:
		numTargets := iter.Varint()
		f(sb, iter.Position(), "num targets", numTargets)
		for i := 0; i < numTargets; i++ {
			f(sb, iter.Position(), "break depth", iter.Varint())
		}
		f(sb, iter.Position(), "break depth for default", iter.Varint())

	case opcodes.MemoryFill, opcodes.MemorySize, opcodes.MemoryGrow:
		f(sb, iter.Position(), "memidx", iter.Varint())

	case opcodes.MemoryCopy:
		f(sb, iter.Position(), "dst memidx", iter.Varint())
		f(sb, iter.Position(), "src memidx", iter.Varint())

	case opcodes.If:
		f(sb, iter.Position(), typeToString(iter.Peek()), iter.Byte())

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
		return fmt.Errorf("unknown opcode: %02x\n", opcode)
	}
	return nil
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
	case 0x70:
		return "funcref"
	default:
		panic(fmt.Sprintf("unknown export kind: %02x", b))
	}
}

func f(sb *strings.Builder, offset int, label string, x any) {
	str := fmt.Sprintf("%s: %s", p(offset), valueToHex(x))
	sb.WriteString(fmt.Sprintf("%s; %s\n", pad(str, 50), label))
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
	case opcodes.Opcode:
		return valueToHex(uint16(x.(opcodes.Opcode)))
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
