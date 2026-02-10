package module

import (
	"fmt"
	"wasp/wasp/internal/funcs"
	"wasp/wasp/internal/instructions"
	"wasp/wasp/internal/iterator"
)

const (
	wasmBinaryMagic   = 0x6d736100
	wasmBinaryVersion = 0x1

	sectionType     = 0x1
	sectionImport   = 0x2
	sectionFunction = 0x3
	sectionExport   = 0x7
	sectionStart    = 0x8
	sectionCode     = 0xa

	typeFunc = 0x60

	kindFunc = 0x00

	guessSize = 0x0
)

func Parse(module *Module, data []byte) error {
	iter := iterator.NewIterator(data)

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

func parseSections(module *Module, iter *iterator.Iterator) (err error) {
	for !iter.Done() {
		sectionOpcode := iter.Varint()
		sectionSize := iter.Varint()

		switch sectionOpcode {
		case sectionType:
			err = parseTypeSection(module, iter)
		case sectionFunction:
			err = parseFunctionSection(module, iter)
		case sectionExport:
			err = parseExportSection(module, iter)
		case sectionCode:
			err = parseCodeSection(module, iter)
		case sectionImport:
			err = parseImportSection(module, iter)
		case sectionStart:
			err = parseStartSection(module, iter)
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

func parseStartSection(module *Module, iter *iterator.Iterator) error {
	module.startFuncIndex = iter.Varint()
	return nil
}

func parseImportSection(module *Module, iter *iterator.Iterator) error {
	numImports := iter.Varint()
	for i := 0; i < numImports; i++ {
		moduleNameLen := iter.Varint()
		moduleName := iter.String(moduleNameLen)
		fieldNameLen := iter.Varint()
		fieldName := iter.String(fieldNameLen)
		importKind := iter.Byte()
		importSignatureIndex := iter.Varint()

		if importKind != kindFunc {
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

func parseTypeSection(module *Module, iter *iterator.Iterator) (err error) {
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

func parseFuncType(module *Module, iter *iterator.Iterator) error {
	numParams := iter.Varint()
	params := make([]int, numParams)
	for i := 0; i < numParams; i++ {
		params[i] = iter.Varint()
	}

	numResults := iter.Varint()
	results := make([]int, numResults)
	for i := 0; i < numResults; i++ {
		results[i] = iter.Varint()
	}

	module.functionSignatures = append(module.functionSignatures, funcs.Signature{
		Params:  params,
		Results: results,
	})

	return nil
}

func parseFunctionSection(module *Module, iter *iterator.Iterator) error {
	numFunctions := iter.Varint()
	for i := 0; i < numFunctions; i++ {
		funcSignatureIndex := iter.Varint()
		module.functions = append(module.functions, funcs.Function{
			Signature: module.functionSignatures[funcSignatureIndex],
		})
	}
	return nil
}

func parseExportSection(module *Module, iter *iterator.Iterator) error {
	numExports := iter.Varint()
	for i := 0; i < numExports; i++ {
		nameLen := iter.Varint()
		name := iter.String(nameLen)
		exportKind := iter.Varint()
		exportIndex := iter.Varint()

		if exportKind != kindFunc {
			return fmt.Errorf("unsupported export kind: 0x%x", exportKind)
		}

		module.exports[name] = Export{
			kind:  exportKind,
			index: exportIndex,
		}
	}
	return nil
}

func parseCodeSection(module *Module, iter *iterator.Iterator) (err error) {
	numFunctions := iter.Varint()
	for i := 0; i < numFunctions; i++ {
		bodySize := iter.Varint()

		var body []byte
		if bodySize == guessSize {
			body, err = iter.ReadUntil(instructions.End.Opcode) // read until end opcode
			if err != nil {
				return fmt.Errorf("failed to read function body: %w", err)
			}
			bodySize = iter.Varint()
		} else {
			body = iter.Bytes(bodySize)
		}

		module.functions[i].Body = iterator.NewIterator(body)
	}
	return nil
}
