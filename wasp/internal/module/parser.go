package module

import (
	"fmt"
	"os"
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/funcs"
	"wasp/wasp/internal/memory"
	"wasp/wasp/internal/opcodes"
	"wasp/wasp/internal/types"
)

var unhandled = make(map[byte]int) // TODO: remove before flight

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
			iter.Assert(opcodes.I32Const)
			offset = iter.Varint()
			iter.Assert(opcodes.End)
		case 0x01: // passive segment
			// Passive segments have no memory index or offset
		case 0x02: // active segment with explicit memory index
			memoryIndex = iter.Varint()
			iter.Assert(opcodes.I32Const)
			offset = iter.Varint()
			iter.Assert(opcodes.End)
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

		iter.Assert(opcodes.I32Const)
		offset := iter.Varint()
		iter.Assert(opcodes.End)

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
		iter.Assert(opcodes.I32Const)
		value := globalType.Read(iter)

		module.globals.Push(value, isMutable)
		iter.Assert(opcodes.End)
	}

	return nil
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
	var offset int

	defer func() {
		if r := recover(); r != nil {
			for opcode := range unhandled {
				fmt.Printf("did not skip immediate for opcode: %x\n", opcode)
			}
			fmt.Printf("Panic: %v\n", r)
			fmt.Printf("Function index: %d\n", index)
			fmt.Printf("Function offset: %x\n", offset)
			fmt.Printf("Iter offset: %x\n", iter.Position())
			os.Exit(1)
		}
	}()

	bodySize := iter.Varint()

	offset = iter.Position()
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
	bodySize -= iter.Position() - offset // adjust body size after reading local decls

	offset = iter.Position() // update offset to point to start of function body

	var body []byte
	if bodySize == guessSize {
		body, err = iter.ReadUntil(opcodes.End) // read until end opcode
		if err != nil {
			return fmt.Errorf("failed to read function body: %w", err)
		}
	} else {
		body = iter.Bytes(bodySize)
	}

	// Precompute block targets
	blocks := precomputeBlocks(body)

	module.functions[index].Locals = locals
	module.functions[index].Body = body
	module.functions[index].Offset = offset
	module.functions[index].Blocks = blocks
	return nil
}

// precomputeBlocks scans the function body and computes block target positions
func precomputeBlocks(body []byte) map[int]funcs.BlockTarget {
	blocks := make(map[int]funcs.BlockTarget)
	bodyIter := binary.NewIterator(body)

	// Stack to track nested blocks during scanning
	type blockEntry struct {
		kind      funcs.BlockKind
		startPos  int
		blockType byte
		elsePos   int
	}
	var stack []blockEntry

	for bodyIter.HasNext() {
		b := bodyIter.Byte()

		switch b {
		case opcodes.Block:
			blockType := bodyIter.Byte()
			stack = append(stack, blockEntry{
				kind:      funcs.BlockKindBlock,
				startPos:  bodyIter.Position(),
				blockType: blockType,
			})

		case opcodes.Loop:
			blockType := bodyIter.Byte()
			stack = append(stack, blockEntry{
				kind:      funcs.BlockKindLoop,
				startPos:  bodyIter.Position(),
				blockType: blockType,
			})

		case opcodes.If:
			blockType := bodyIter.Byte()
			stack = append(stack, blockEntry{
				kind:      funcs.BlockKindIf,
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
				blocks[entry.startPos] = funcs.BlockTarget{
					Kind:      entry.kind,
					StartPos:  entry.startPos,
					ElsePos:   entry.elsePos,
					EndPos:    bodyIter.Position(),
					BlockType: entry.blockType,
				}
			}

		default:
			// Skip immediates for other instructions
			skipParseImmediates(bodyIter, b)
		}
	}

	return blocks
}

// skipParseImmediates skips instruction immediates during block precomputation
func skipParseImmediates(iter *binary.Iterator, opcode byte) {
	switch opcode {
	// Control flow with immediates
	case opcodes.Br, opcodes.BrIf:
		iter.Varint() // label index
	case opcodes.BrTable:
		numLabels := iter.Varint() // vector length
		for i := 0; i < numLabels; i++ {
			iter.Varint() // label index
		}
		iter.Varint() // default label
	case opcodes.Call:
		iter.Varint() // function index
	case opcodes.CallIndirect:
		iter.Varint() // type index
		iter.Varint() // table index
	case opcodes.ReturnCallIndirect:
		iter.Varint() // type index
		iter.Varint() // table index

	// Variable instructions with immediates
	case opcodes.LocalGet, opcodes.LocalSet, opcodes.LocalTee:
		iter.Varint() // local index
	case opcodes.GlobalGet, opcodes.GlobalSet:
		iter.Varint() // global index

	// Const instructions with immediates
	case opcodes.I32Const:
		iter.Varint() // i32 value
	case opcodes.I64Const:
		iter.Varint() // i64 value
	case opcodes.F32Const:
		iter.Bytes(4) // f32 value
	case opcodes.F64Const:
		iter.Bytes(8) // f64 value

	// Memory instructions with immediates
	case opcodes.I32Load, opcodes.I32Store,
		opcodes.I64Load, opcodes.F32Load, opcodes.F64Load,
		opcodes.I32Load8S, opcodes.I32Load8U, opcodes.I32Load16S, opcodes.I32Load16U,
		opcodes.I64Load8S, opcodes.I64Load8U, opcodes.I64Load16S, opcodes.I64Load16U,
		opcodes.I64Load32S, opcodes.I64Load32U,
		opcodes.I64Store, opcodes.F32Store, opcodes.F64Store,
		opcodes.I32Store8, opcodes.I32Store16,
		opcodes.I64Store8, opcodes.I64Store16, opcodes.I64Store32:
		iter.Varint() // align
		iter.Varint() // offset
	case opcodes.MemorySize, opcodes.MemoryGrow:
		iter.Byte() // memory index

	// FC extensions (0xFC prefix)
	case 0xfc:
		subOpcode := iter.Varint()
		switch subOpcode {
		case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07:
			// i32.trunc_sat_f32_s/u, i32.trunc_sat_f64_s/u,
			// i64.trunc_sat_f32_s/u, i64.trunc_sat_f64_s/u
			// no additional immediates
		case 0x08: // memory.init
			iter.Varint() // data index
			iter.Byte()   // memory index (always 0)
		case 0x09: // data.drop
			iter.Varint() // data index
		case 0x0a: // memory.copy
			iter.Byte() // dest memory index
			iter.Byte() // src memory index
		case 0x0b: // memory.fill
			iter.Byte() // memory index
		case 0x0c: // table.init
			iter.Varint() // elem index
			iter.Varint() // table index
		case 0x0d: // elem.drop
			iter.Varint() // elem index
		case 0x0e: // table.copy
			iter.Varint() // dest table index
			iter.Varint() // src table index
		case 0x0f: // table.grow
			iter.Varint() // table index
		case 0x10: // table.size
			iter.Varint() // table index
		case 0x11: // table.fill
			iter.Varint() // table index
		}

	// Instructions with NO immediates (stack-only operations)
	case opcodes.Unreachable, opcodes.Nop, opcodes.Return, opcodes.Drop, opcodes.Select:
		// no immediates
	case opcodes.I32Eqz, opcodes.I32Eq, opcodes.I32Ne,
		opcodes.I32LtS, opcodes.I32LtU, opcodes.I32GtS, opcodes.I32GtU,
		opcodes.I32LeS, opcodes.I32LeU, opcodes.I32GeS, opcodes.I32GeU:
		// i32 comparison - no immediates
	case opcodes.I64Eqz, opcodes.I64Eq, opcodes.I64Ne,
		opcodes.I64LtS, opcodes.I64LtU, opcodes.I64GtS, opcodes.I64GtU,
		opcodes.I64LeS, opcodes.I64LeU, opcodes.I64GeS, opcodes.I64GeU:
		// i64 comparison - no immediates
	case opcodes.F32Eq, opcodes.F32Ne, opcodes.F32Lt, opcodes.F32Gt, opcodes.F32Le, opcodes.F32Ge:
		// f32 comparison - no immediates
	case opcodes.F64Eq, opcodes.F64Ne, opcodes.F64Lt, opcodes.F64Gt, opcodes.F64Le, opcodes.F64Ge:
		// f64 comparison - no immediates
	case opcodes.I32Clz, opcodes.I32Ctz, opcodes.I32Popcnt,
		opcodes.I32Add, opcodes.I32Sub, opcodes.I32Mul, opcodes.I32DivS, opcodes.I32DivU,
		opcodes.I32RemS, opcodes.I32RemU, opcodes.I32And, opcodes.I32Or, opcodes.I32Xor,
		opcodes.I32Shl, opcodes.I32ShrS, opcodes.I32ShrU, opcodes.I32Rotl, opcodes.I32Rotr:
		// i32 arithmetic - no immediates
	case opcodes.I64Clz, opcodes.I64Ctz, opcodes.I64Popcnt,
		opcodes.I64Add, opcodes.I64Sub, opcodes.I64Mul, opcodes.I64DivS, opcodes.I64DivU,
		opcodes.I64RemS, opcodes.I64RemU, opcodes.I64And, opcodes.I64Or, opcodes.I64Xor,
		opcodes.I64Shl, opcodes.I64ShrS, opcodes.I64ShrU, opcodes.I64Rotl, opcodes.I64Rotr:
		// i64 arithmetic - no immediates
	case opcodes.F32Abs, opcodes.F32Neg, opcodes.F32Ceil, opcodes.F32Floor, opcodes.F32Trunc,
		opcodes.F32Nearest, opcodes.F32Sqrt, opcodes.F32Add, opcodes.F32Sub, opcodes.F32Mul,
		opcodes.F32Div, opcodes.F32Min, opcodes.F32Max, opcodes.F32Copysign:
		// f32 arithmetic - no immediates
	case opcodes.F64Abs, opcodes.F64Neg, opcodes.F64Ceil, opcodes.F64Floor, opcodes.F64Trunc,
		opcodes.F64Nearest, opcodes.F64Sqrt, opcodes.F64Add, opcodes.F64Sub, opcodes.F64Mul,
		opcodes.F64Div, opcodes.F64Min, opcodes.F64Max, opcodes.F64Copysign:
		// f64 arithmetic - no immediates
	case opcodes.I32WrapI64, opcodes.I32TruncF32S, opcodes.I32TruncF32U,
		opcodes.I32TruncF64S, opcodes.I32TruncF64U, opcodes.I64ExtendI32S, opcodes.I64ExtendI32U,
		opcodes.I64TruncF32S, opcodes.I64TruncF32U, opcodes.I64TruncF64S, opcodes.I64TruncF64U,
		opcodes.F32ConvertI32S, opcodes.F32ConvertI32U, opcodes.F32ConvertI64S, opcodes.F32ConvertI64U,
		opcodes.F32DemoteF64, opcodes.F64ConvertI32S, opcodes.F64ConvertI32U, opcodes.F64ConvertI64S,
		opcodes.F64ConvertI64U, opcodes.F64PromoteF32:
		// conversion - no immediates
	case opcodes.I32ReinterpretF32, opcodes.I64ReinterpretF64,
		opcodes.F32ReinterpretI32, opcodes.F64ReinterpretI64:
		// reinterpret - no immediates

	default:
		unhandled[opcode]++
	}
}
