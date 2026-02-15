package module

import (
	"fmt"
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/funcs"
	"wasp/wasp/internal/funcs/fnblock"
	"wasp/wasp/internal/memory"
	"wasp/wasp/internal/opcodes"
	"wasp/wasp/internal/types"
)

var unhandled = make(map[uint16]int) // TODO: remove before flight

const (
	wasmBinaryMagic   = 0x6d736100
	wasmBinaryVersion = 0x1

	sectionCustom    = 0x0
	sectionType      = 0x1
	sectionImport    = 0x2
	sectionFunction  = 0x3
	sectionTable     = 0x4
	sectionMemory    = 0x5
	sectionGlobal    = 0x6
	sectionExport    = 0x7
	sectionStart     = 0x8
	sectionElement   = 0x9
	sectionCode      = 0xa
	sectionData      = 0xb
	sectionDataCount = 0xc
	sectionTag       = 0xd

	typeFunc = 0x60

	exportKindFunc   = 0x00
	exportKindMemory = 0x02

	guessSize = 0x0

	funcRef = 0x70
)

func Parse(module *Module, data []byte) error {
	iter := binary.NewIterator(data)

	binaryMagic := iter.Uint32()
	if binaryMagic != wasmBinaryMagic {
		return fmt.Errorf("invalid wasm binary magic: 0x%x", binaryMagic)
	}

	version := iter.Uint32()
	if version != wasmBinaryVersion {
		return fmt.Errorf("unsupported wasm version: %d", version)
	}

	if err := parseSections(module, iter); err != nil {
		return fmt.Errorf("failed to parse wasm sections: %w", err)
	}

	return nil
}

func parseSections(module *Module, iter *binary.Iterator) (err error) {
	for iter.HasNext() {
		sectionOpcode := iter.Varint()
		sectionSize := iter.Varint()

		switch sectionOpcode {
		case sectionCustom:
			err = parseCustomSection(module, iter, sectionSize)
		case sectionType:
			err = parseTypeSection(module, iter)
		case sectionImport:
			err = parseImportSection(module, iter)
		case sectionFunction:
			err = parseFunctionSection(module, iter)
		case sectionTable:
			err = parseTableSection(module, iter)
		case sectionMemory:
			err = parseMemorySection(module, iter)
		case sectionGlobal:
			err = parseGlobalSection(module, iter)
		case sectionExport:
			err = parseExportSection(module, iter)
		case sectionStart:
			err = parseStartSection(module, iter)
		case sectionElement:
			err = parseElementSection(module, iter)
		case sectionCode:
			err = parseCodeSection(module, iter)
		case sectionData:
			err = parseDataSection(module, iter)
		case sectionDataCount:
			err = parseDataCountSection(module, iter)
		default:
			return fmt.Errorf("invalid section type: 0x%x", sectionOpcode)
		}

		if err != nil {
			return fmt.Errorf("failed to parse section 0x%x: %w", sectionOpcode, err)
		}

		if sectionSize == guessSize {
			sectionSize = iter.Varint()
		}
	}
	return nil
}

func parseCustomSection(module *Module, iter *binary.Iterator, sectionSize int) error {
	startPos := iter.Position()

	nameLen := iter.Varint()
	name := iter.String(nameLen)

	// Calculate remaining bytes in the section
	bytesRead := iter.Position() - startPos
	dataLen := sectionSize - bytesRead
	data := iter.Bytes(dataLen)

	module.customSections = append(module.customSections, CustomSection{
		Name: name,
		Data: data,
	})
	return nil
}

func parseDataSection(module *Module, iter *binary.Iterator) error {
	numDataSegments := iter.Varint()
	for i := 0; i < numDataSegments; i++ {
		segmentFlags := iter.Byte()

		var memoryIndex int
		var offset int

		switch segmentFlags {
		case 0x00: // active segment with memory index 0
			memoryIndex = 0
			assertOpcode(iter, opcodes.I32Const)
			offset = iter.Varint()
			assertOpcode(iter, opcodes.End)
		case 0x01: // passive segment
			// Passive segments have no memory index or offset
		case 0x02: // active segment with explicit memory index
			memoryIndex = iter.Varint()
			assertOpcode(iter, opcodes.I32Const)
			offset = iter.Varint()
			assertOpcode(iter, opcodes.End)
		default:
			return fmt.Errorf("unsupported data segment flag: 0x%x", segmentFlags)
		}

		dataLen := iter.Varint()
		data := iter.Bytes(dataLen)

		module.data = append(module.data, DataSegment{
			MemoryIndex: memoryIndex,
			Offset:      offset,
			Data:        data,
		})
	}
	return nil
}

func parseDataCountSection(module *Module, iter *binary.Iterator) error {
	dataCount := iter.Varint()
	_ = dataCount // TODO: store data count info in module struct instead of ignoring
	return nil
}

func parseElementSection(module *Module, iter *binary.Iterator) error {
	numElementSegments := iter.Varint()
	for i := 0; i < numElementSegments; i++ {
		segmentFlags := iter.Byte()
		if segmentFlags != 0x0 {
			return fmt.Errorf("unsupported element segment flag: 0x%x", segmentFlags)
		}

		assertOpcode(iter, opcodes.I32Const)
		offset := iter.Varint()
		assertOpcode(iter, opcodes.End)

		_ = offset // TODO: store offset info in module struct instead of ignoring

		numElements := iter.Varint()
		for j := 0; j < numElements; j++ {
			elementFuncIndex := iter.Varint()
			_ = elementFuncIndex // TODO: store element info in module struct instead of ignoring
		}
	}
	return nil
}

func parseMemorySection(module *Module, iter *binary.Iterator) error {
	numMemories := iter.Varint()
	module.memories = make([]*memory.Memory, numMemories)
	for i := 0; i < numMemories; i++ {
		flag := iter.Byte()
		initialPages := iter.Varint()
		var maxPages int
		if flag == 0x1 {
			maxPages = iter.Varint()
		}
		module.memories[i] = memory.NewMemory(initialPages, maxPages)
	}
	return nil
}

func parseTableSection(module *Module, iter *binary.Iterator) error {
	numTables := iter.Varint()
	for i := 0; i < numTables; i++ {
		elementType := iter.Byte()
		if elementType != funcRef {
			return fmt.Errorf("unsupported table element type: 0x%x", elementType)
		}

		flag := iter.Byte()
		initialSize := iter.Varint()
		var maxSize int
		if flag == 0x1 {
			maxSize = iter.Varint()
		}

		module.tables = append(module.tables, Table{
			ElementType: elementType,
			InitialSize: initialSize,
			MaxSize:     maxSize,
		})
	}
	return nil
}

func parseGlobalSection(module *Module, iter *binary.Iterator) error {
	numGlobals := iter.Varint()
	for i := 0; i < numGlobals; i++ {
		globalType := types.ForCode(iter.Byte())
		isMutable := iter.BoolByte()
		assertOpcode(iter, opcodes.I32Const)
		value := globalType.Read(iter)

		module.globals.Push(value, isMutable)
		assertOpcode(iter, opcodes.End)
	}

	return nil
}

func assertOpcode(iter *binary.Iterator, expected opcodes.Opcode) {
	opcode := iter.Opcode()
	if opcode != expected {
		panic(fmt.Sprintf("assertion failed: expected bytes %v, got %v", expected, opcode))
	}
}

func parseStartSection(module *Module, iter *binary.Iterator) error {
	module.startFuncIndex = iter.Varint()
	return nil
}

func parseImportSection(module *Module, iter *binary.Iterator) error {
	numImports := iter.Varint()
	for i := 0; i < numImports; i++ {
		moduleNameLen := iter.Varint()
		moduleName := iter.String(moduleNameLen)
		fieldNameLen := iter.Varint()
		fieldName := iter.String(fieldNameLen)
		importKind := iter.Byte()
		importSignatureIndex := iter.Varint()

		if importKind != exportKindFunc { // TODO: check kind
			return fmt.Errorf("invalid import kind 0x%x", importKind)
		}

		module.imports = append(module.imports, Import{
			ModuleName: moduleName,
			FieldName:  fieldName,
			Kind:       importKind,
			Signature:  module.functionSignatures[importSignatureIndex],
		})
	}
	return nil
}

func parseTypeSection(module *Module, iter *binary.Iterator) (err error) {
	numTypes := iter.Varint()
	for i := 0; i < numTypes; i++ {
		typeCode := iter.Varint()

		switch typeCode {
		case typeFunc:
			err = parseFuncType(module, iter)
		default:
			return fmt.Errorf("invalid type code: 0x%x", typeCode)
		}

		if err != nil {
			return fmt.Errorf("failed to parse type %d: %w", i, err)
		}
	}
	return nil
}

func parseFuncType(module *Module, iter *binary.Iterator) error {
	numParams := iter.Varint()
	params := make([]byte, numParams)
	for i := 0; i < numParams; i++ {
		params[i] = iter.Byte()
	}

	numResults := iter.Varint()
	results := make([]byte, numResults)
	for i := 0; i < numResults; i++ {
		results[i] = iter.Byte()
	}

	module.functionSignatures = append(module.functionSignatures, funcs.Signature{
		Params:  params,
		Results: results,
	})

	return nil
}

func parseFunctionSection(module *Module, iter *binary.Iterator) error {
	numFunctions := iter.Varint()
	module.functions = make([]funcs.Function, numFunctions)
	for i := 0; i < numFunctions; i++ {
		funcSignatureIndex := iter.Varint()
		module.functions[i] = funcs.Function{
			Index:     i,
			Signature: module.functionSignatures[funcSignatureIndex],
		}
	}
	return nil
}

func parseExportSection(module *Module, iter *binary.Iterator) error {
	numExports := iter.Varint()
	for i := 0; i < numExports; i++ {
		nameLen := iter.Varint()
		name := iter.String(nameLen)
		exportKind := iter.Varint() // TODO: check kind
		exportIndex := iter.Varint()

		module.exports[name] = Export{
			kind:  exportKind,
			index: exportIndex,
		}
	}
	return nil
}

func parseCodeSection(module *Module, iter *binary.Iterator) (err error) {
	numFunctions := iter.Varint()
	for i := 0; i < numFunctions; i++ {
		if err = parseFunction(module, iter, i); err != nil {
			return fmt.Errorf("failed to parse function %d: %w", i, err)
		}
	}
	return nil
}

func parseFunction(module *Module, iter *binary.Iterator, index int) (err error) {
	var funcOffset int

	bodySize := iter.Varint()

	funcOffset = iter.Position()
	localDeclCount := iter.Varint()
	locals := make([]any, 0, localDeclCount)
	for j := 0; j < localDeclCount; j++ {
		localTypeCount := iter.Varint()
		typeCode := iter.Byte()
		localType := types.ForCode(typeCode)
		for k := 0; k < localTypeCount; k++ {
			locals = append(locals, localType.Zero())
		}
	}

	bodySize -= iter.Position() - funcOffset // adjust body size after reading local decls

	bodyOffset := iter.Position() // update offset to point to start of function body

	var body []byte
	if bodySize == guessSize {
		body, err = iter.ReadUntil(byte(opcodes.End)) // read until end opcode
		if err != nil {
			return fmt.Errorf("failed to read function body: %w", err)
		}
	} else {
		body = iter.Bytes(bodySize)
	}

	// Precompute block targets
	blocks, err := precomputeBlocks(body)
	if err != nil {
		Foo(
			iter.Range(funcOffset, iter.Position()),
			index,
			funcOffset,
		)
		return fmt.Errorf("failed to read function blocks: %w", err)
	}

	module.functions[index].Locals = locals
	module.functions[index].Body = body
	module.functions[index].Offset = bodyOffset
	module.functions[index].Blocks = blocks
	return nil
}

// precomputeBlocks scans the function body and computes block target positions
func precomputeBlocks(body []byte) (map[int]fnblock.Target, error) {
	blocks := make(map[int]fnblock.Target)
	bodyIter := binary.NewIterator(body)

	// Stack to track nested fnblock during scanning
	type blockEntry struct {
		kind      fnblock.Kind
		startPos  int
		blockType byte
		elsePos   int
	}
	var stack []blockEntry

	for bodyIter.HasNext() {
		b := bodyIter.Opcode()

		switch b {
		case opcodes.Block:
			blockType := bodyIter.Byte()
			stack = append(stack, blockEntry{
				kind:      fnblock.KindBlock,
				startPos:  bodyIter.Position(),
				blockType: blockType,
			})

		case opcodes.Loop:
			blockType := bodyIter.Byte()
			stack = append(stack, blockEntry{
				kind:      fnblock.KindLoop,
				startPos:  bodyIter.Position(),
				blockType: blockType,
			})

		case opcodes.If:
			blockType := bodyIter.Byte()
			stack = append(stack, blockEntry{
				kind:      fnblock.KindIf,
				startPos:  bodyIter.Position(),
				blockType: blockType,
			})

		case opcodes.Else:
			if len(stack) > 0 {
				stack[len(stack)-1].elsePos = bodyIter.Position()
			}

		case opcodes.End:
			if len(stack) > 0 {
				entry := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				blocks[entry.startPos] = fnblock.Target{
					Kind:      entry.kind,
					StartPos:  entry.startPos,
					ElsePos:   entry.elsePos,
					EndPos:    bodyIter.Position(),
					BlockType: entry.blockType,
				}
			}

		default:
			// Skip immediates for other instructions
			if err := skipParseImmediates(bodyIter, b); err != nil {
				return nil, err
			}
		}
	}

	return blocks, nil
}

// skipParseImmediates skips instruction immediates during block precomputation
func skipParseImmediates(iter *binary.Iterator, opcode opcodes.Opcode) error {
	switch opcode {
	case opcodes.GlobalGet, opcodes.GlobalSet:
		iter.Varint() // global index
	case opcodes.I32Const:
		iter.Varint() // i32 literal
	case opcodes.I64Const:
		iter.Varint() // i64 literal
	case opcodes.F64Const:
		iter.Float64() // f64 literal
	case opcodes.F32Const:
		iter.Float32() // f32 literal
	case opcodes.LocalTee, opcodes.LocalGet, opcodes.LocalSet:
		iter.Varint() // local index
	case opcodes.Block, opcodes.Loop:
		iter.Byte() // block type
	case opcodes.Br, opcodes.BrIf:
		iter.Varint() // break depth
	case opcodes.I32Load8U, opcodes.I32Load8S, opcodes.I32Load,
		opcodes.I64Load, opcodes.I32Load16S, opcodes.I32Load16U,
		opcodes.I64Load8S, opcodes.I64Load8U, opcodes.I64Load16S,
		opcodes.I64Load16U, opcodes.I64Load32S, opcodes.I64Load32U,
		opcodes.F64Load, opcodes.F32Load:
		iter.Byte()   // alignment
		iter.Varint() // load offset
	case opcodes.I32Store, opcodes.I64Store, opcodes.I32Store16,
		opcodes.I32Store8, opcodes.I64Store16, opcodes.I64Store8,
		opcodes.I64Store32, opcodes.F64Store, opcodes.F32Store:
		iter.Byte()   // alignment
		iter.Varint() // store offset
	case opcodes.Call, opcodes.ReturnCall:
		iter.Varint() // function index
	case opcodes.CallIndirect:
		iter.Varint() // signature index
		iter.Varint() // table index
	case opcodes.BrTable:
		numTargets := iter.Varint()
		for i := 0; i < numTargets; i++ {
			iter.Varint() // break depth
		}
		iter.Varint() // break depth for default
	case opcodes.MemoryFill, opcodes.MemorySize, opcodes.MemoryGrow:
		iter.Varint() // memidx
	case opcodes.MemoryCopy:
		iter.Varint() // dst memidx
		iter.Varint() // src memidx
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
		return nil
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
		return nil
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
		return nil
	case opcodes.F32ConvertI32S, opcodes.F32ConvertI32U, opcodes.F32ConvertI64S,
		opcodes.F32ConvertI64U, opcodes.F32DemoteF64, opcodes.F32Mul,
		opcodes.F32Add, opcodes.F32Sub, opcodes.F32Div, opcodes.F32Min,
		opcodes.F32Max, opcodes.F32Copysign, opcodes.F32Sqrt,
		opcodes.F32Ceil, opcodes.F32Floor, opcodes.F32Trunc,
		opcodes.F32Nearest, opcodes.F32ReinterpretI32,
		opcodes.F32Abs, opcodes.F32Lt, opcodes.F32Gt, opcodes.F32Le,
		opcodes.F32Ge, opcodes.F32Eq, opcodes.F32Ne:
		// no immediate arguments
		return nil
	case opcodes.End, opcodes.Drop, opcodes.Select,
		opcodes.Unreachable, opcodes.Return:
		// no immediate arguments
		return nil
	default:
		return fmt.Errorf("unknown opcode: %02x\n", opcode)
	}
	return nil
}
