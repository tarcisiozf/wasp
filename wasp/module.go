package wasp

import (
	"fmt"
	"os"
)

const (
	wasmBinaryMagic   = 0x6d736100
	wasmBinaryVersion = 0x1

	sectionType     = 0x1
	sectionFunction = 0x3
	sectionExport   = 0x7
	sectionCode     = 0xa

	typeFunc = 0x60

	exportKindFunc = 0x00

	guessSize = 0x0
)

type Function struct {
	params  []byte
	results []byte
	body    []byte
}

type Export struct {
	kind  byte
	index byte
}

type Module struct {
	functions []Function
	exports   map[string]Export
}

func NewModule(binary []byte) (*Module, error) {
	module := &Module{
		exports: make(map[string]Export),
	}
	if err := module.parse(binary); err != nil {
		return nil, fmt.Errorf("failed to parse wasm module: %w", err)
	}

	return module, nil
}

func NewModuleFromFile(path string) (*Module, error) {
	binary, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read wasm file: %w", err)
	}
	return NewModule(binary)
}

func (module *Module) parse(data []byte) error {
	iter := NewIterator(data)

	binaryMagic := iter.uint32()
	if binaryMagic != wasmBinaryMagic {
		return fmt.Errorf("invalid wasm binary magic: 0x%x", binaryMagic)
	}

	version := iter.uint32()
	if version != wasmBinaryVersion {
		return fmt.Errorf("unsupported wasm version: %d", version)
	}

	if err := module.parseSections(iter); err != nil {
		return fmt.Errorf("failed to parse wasm sections: %w", err)
	}

	return nil
}

func (module *Module) parseSections(iter *Iterator) (err error) {
	for !iter.done() {
		sectionOpcode := iter.byte()
		sectionSize := iter.byte()

		switch sectionOpcode {
		case sectionType:
			err = module.parseTypeSection(iter)
		case sectionFunction:
			err = module.parseFunctionSection(iter)
		case sectionExport:
			err = module.parseExportSection(iter)
		case sectionCode:
			err = module.parseCodeSection(iter)
		default:
			return fmt.Errorf("invalid section type: 0x%x", sectionOpcode)
		}

		if err != nil {
			return fmt.Errorf("failed to parse section 0x%x: %w", sectionOpcode, err)
		}

		if sectionSize == guessSize {
			sectionSize = iter.byte()
		}
	}
	return nil
}

func (module *Module) parseTypeSection(iter *Iterator) (err error) {
	numTypes := iter.byte()
	for i := 0; i < int(numTypes); i++ {
		typeCode := iter.byte()

		switch typeCode {
		case typeFunc:
			err = module.parseFuncType(iter)
		default:
			return fmt.Errorf("invalid type code: 0x%x", typeCode)
		}

		if err != nil {
			return fmt.Errorf("failed to parse type %d: %w", i, err)
		}
	}
	return nil
}

func (module *Module) parseFuncType(iter *Iterator) error {
	numParams := iter.byte()
	params := make([]byte, numParams)
	for i := 0; i < int(numParams); i++ {
		params[i] = iter.byte()
	}

	numResults := iter.byte()
	results := make([]byte, numResults)
	for i := 0; i < int(numResults); i++ {
		results[i] = iter.byte()
	}

	module.addFunction(params, results)

	return nil
}

func (module *Module) parseFunctionSection(iter *Iterator) error {
	numFunctions := iter.byte()
	for i := 0; i < int(numFunctions); i++ {
		_ = iter.byte() // func signature index
	}
	return nil
}

func (module *Module) parseExportSection(iter *Iterator) error {
	numExports := iter.byte()
	for i := 0; i < int(numExports); i++ {
		nameLen := iter.byte()
		name := string(iter.bytes(int(nameLen)))
		exportKind := iter.byte()
		exportIndex := iter.byte()

		if exportKind != 0x00 {
			return fmt.Errorf("unsupported export kind: 0x%x", exportKind)
		}

		module.exports[name] = Export{
			kind:  exportKind,
			index: exportIndex,
		}
	}
	return nil
}

func (module *Module) parseCodeSection(iter *Iterator) (err error) {
	numFunctions := iter.byte()
	for i := 0; i < int(numFunctions); i++ {
		bodySize := iter.byte()

		var body []byte
		if bodySize == guessSize {
			body, err = iter.readUntil(0x0b) // read until end opcode
			if err != nil {
				return fmt.Errorf("failed to read function body: %w", err)
			}
			bodySize = iter.byte()
		} else {
			body = iter.bytes(int(bodySize))
		}

		module.functions[i].body = body
	}
	return nil
}

func (module *Module) addFunction(params []byte, results []byte) int {
	index := len(module.functions)
	module.functions = append(module.functions, Function{
		params:  params,
		results: results,
	})
	return index
}
