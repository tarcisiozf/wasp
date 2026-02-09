package wasp

import "fmt"

func parseModule(module *Module, data []byte) error {
	iter := newIterator(data)

	binaryMagic := iter.uint32()
	if binaryMagic != wasmBinaryMagic {
		return fmt.Errorf("invalid wasm binary magic: 0x%x", binaryMagic)
	}

	version := iter.uint32()
	if version != wasmBinaryVersion {
		return fmt.Errorf("unsupported wasm version: %d", version)
	}

	if err := parseSections(module, iter); err != nil {
		return fmt.Errorf("failed to parse wasm sections: %w", err)
	}

	return nil
}

func parseSections(module *Module, iter *Iterator) (err error) {
	for !iter.done() {
		sectionOpcode := iter.varint()
		sectionSize := iter.varint()

		switch sectionOpcode {
		case sectionType:
			err = parseTypeSection(module, iter)
		case sectionFunction:
			err = parseFunctionSection(module, iter)
		case sectionExport:
			err = parseExportSection(module, iter)
		case sectionCode:
			err = parseCodeSection(module, iter)
		default:
			return fmt.Errorf("invalid section type: 0x%x", sectionOpcode)
		}

		if err != nil {
			return fmt.Errorf("failed to parse section 0x%x: %w", sectionOpcode, err)
		}

		if sectionSize == guessSize {
			sectionSize = iter.varint()
		}
	}
	return nil
}

func parseTypeSection(module *Module, iter *Iterator) (err error) {
	numTypes := iter.varint()
	for i := 0; i < numTypes; i++ {
		typeCode := iter.varint()

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

func parseFuncType(module *Module, iter *Iterator) error {
	numParams := iter.varint()
	params := make([]int, numParams)
	for i := 0; i < numParams; i++ {
		params[i] = iter.varint()
	}

	numResults := iter.varint()
	results := make([]int, numResults)
	for i := 0; i < numResults; i++ {
		results[i] = iter.varint()
	}

	module.addFunction(params, results)

	return nil
}

func parseFunctionSection(module *Module, iter *Iterator) error {
	numFunctions := iter.varint()
	for i := 0; i < numFunctions; i++ {
		_ = iter.varint() // func signature index
	}
	return nil
}

func parseExportSection(module *Module, iter *Iterator) error {
	numExports := iter.varint()
	for i := 0; i < numExports; i++ {
		nameLen := iter.varint()
		name := string(iter.bytes(nameLen))
		exportKind := iter.varint()
		exportIndex := iter.varint()

		if exportKind != exportKindFunc {
			return fmt.Errorf("unsupported export kind: 0x%x", exportKind)
		}

		module.exports[name] = Export{
			kind:  exportKind,
			index: exportIndex,
		}
	}
	return nil
}

func parseCodeSection(module *Module, iter *Iterator) (err error) {
	numFunctions := iter.varint()
	for i := 0; i < numFunctions; i++ {
		bodySize := iter.varint()

		var body []byte
		if bodySize == guessSize {
			body, err = iter.readUntil(opcodeEnd) // read until end opcode
			if err != nil {
				return fmt.Errorf("failed to read function body: %w", err)
			}
			bodySize = iter.varint()
		} else {
			body = iter.bytes(bodySize)
		}

		module.functions[i].body = newIterator(body)
	}
	return nil
}
