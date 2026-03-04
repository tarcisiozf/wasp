package wasp

import (
	"fmt"
	"io"

	"github.com/tarcisiozf/wasp/internal/binary"
	"github.com/tarcisiozf/wasp/internal/binary/leb"
	"github.com/tarcisiozf/wasp/internal/execution"
	"github.com/tarcisiozf/wasp/internal/memory"
	iface "github.com/tarcisiozf/wasp/memory"
)

type StateStore struct {
	Globals  []GlobalItem
	Memories []MemoryState
	Tables   []TableState
}

type GlobalItem struct {
	Value   any
	Mutable bool
}

type MemoryState struct {
	Data     []byte
	NumPages int
	MaxPages int
}

type TableState = memory.Table

type CallFrame struct {
	FunctionIndex int
	Context       CallContext
}

type BlockFrame struct {
	StartPos int
}

type CallContext struct {
	Stack  []any
	Locals []any

	NumParams  int
	NumResults int
	Params     []any

	BodyPos             int
	BodyCheckpoint      int
	FunctionCallRequest int
	Done                bool
	TailCall            bool

	Condition  bool
	BlockStack []BlockFrame
}

type ExecutionState struct {
	Store     StateStore
	CallStack []CallFrame
}

const (
	tagInt   byte = 0x01
	tagBool  byte = 0x02
	tagByte  byte = 0x03
	tagBytes byte = 0x04
	tagInt32 byte = 0x05
	tagInt64 byte = 0x06
)

type Encoder struct {
	dest io.Writer
}

func (enc *Encoder) Int(value int) {
	enc.VarUint(uint64(value))
}

func (enc *Encoder) VarUint(value uint64) {
	var buf [10]byte
	n := leb.EncodeUint(buf[:], value)
	enc.dest.Write(buf[:n])
}

func (enc *Encoder) Any(value any) {
	switch value.(type) {
	case int:
		enc.Byte(tagInt)
		enc.Int(value.(int))
	case bool:
		enc.Byte(tagBool)
		enc.Bool(value.(bool))
	case byte:
		enc.Byte(tagByte)
		enc.Byte(value.(byte))
	case []byte:
		enc.Byte(tagBytes)
		enc.Bytes(value.([]byte))
	case int32:
		enc.Byte(tagInt32)
		enc.VarUint(uint64(value.(int32)))
	case int64:
		enc.Byte(tagInt64)
		enc.VarUint(uint64(value.(int64)))
	default:
		panic(fmt.Sprintf("unsupported encoding for type: %T", value))
	}
}

func (enc *Encoder) Bool(value bool) {
	if value {
		enc.Byte(1)
	} else {
		enc.Byte(0)
	}
}

func (enc *Encoder) Bytes(data []byte) {
	enc.Int(len(data))
	enc.dest.Write(data)
}

func (enc *Encoder) Byte(b byte) {
	enc.dest.Write([]byte{b})
}

func NewEncoder(dest io.Writer) *Encoder {
	return &Encoder{dest: dest}
}

type Decoder struct {
	src io.Reader
}

func NewDecoder(src io.Reader) *Decoder {
	return &Decoder{src: src}
}

func (dec *Decoder) Byte() (byte, error) {
	var buf [1]byte
	if _, err := io.ReadFull(dec.src, buf[:]); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (dec *Decoder) Bool() (bool, error) {
	b, err := dec.Byte()
	if err != nil {
		return false, err
	}
	return b != 0, nil
}

func (dec *Decoder) VarUint() (uint64, error) {
	var x uint64
	var shift uint
	for {
		b, err := dec.Byte()
		if err != nil {
			return 0, err
		}
		x |= uint64(b&0x7F) << shift
		if b < 0x80 {
			break
		}
		shift += 7
	}
	return x, nil
}

func (dec *Decoder) Int() (int, error) {
	v, err := dec.VarUint()
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

func (dec *Decoder) Bytes() ([]byte, error) {
	length, err := dec.Int()
	if err != nil {
		return nil, err
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(dec.src, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (dec *Decoder) Any() (any, error) {
	tag, err := dec.Byte()
	if err != nil {
		return nil, err
	}
	switch tag {
	case tagInt:
		return dec.Int()
	case tagBool:
		return dec.Bool()
	case tagByte:
		return dec.Byte()
	case tagBytes:
		return dec.Bytes()
	case tagInt32:
		v, err := dec.VarUint()
		if err != nil {
			return nil, err
		}
		return int32(v), nil
	case tagInt64:
		v, err := dec.VarUint()
		if err != nil {
			return nil, err
		}
		return int64(v), nil
	default:
		return nil, fmt.Errorf("unsupported type tag: 0x%02x", tag)
	}
}

func SerializeState(dest io.Writer, store *Store, instance *Instance) error {
	encoder := NewEncoder(dest)

	if err := toStateStore(encoder, store); err != nil {
		return err
	}
	if err := encodeCallStack(encoder, instance); err != nil {
		return err
	}

	return nil
}

func DeserializeState(src io.Reader, module *Module, linker *Linker) (*Store, *Instance, error) {
	decoder := NewDecoder(src)

	store, err := fromStateStore(decoder)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode store: %w", err)
	}

	instance, err := NewInstance(module, store, WithLinker(linker))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create instance: %w", err)
	}

	if err := decodeCallStack(decoder, module, instance, store); err != nil {
		return nil, nil, fmt.Errorf("failed to decode call stack: %w", err)
	}

	return store, instance, nil
}

func fromStateStore(decoder *Decoder) (*Store, error) {
	globals, err := decodeGlobalItems(decoder)
	if err != nil {
		return nil, fmt.Errorf("failed to decode globals: %w", err)
	}
	memories, err := decodeMemories(decoder)
	if err != nil {
		return nil, fmt.Errorf("failed to decode memories: %w", err)
	}
	tables, err := decodeTables(decoder)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tables: %w", err)
	}
	return &Store{
		Globals:  globals,
		Memories: memories,
		Tables:   tables,
	}, nil
}

func decodeGlobalItems(decoder *Decoder) (*memory.Global, error) {
	size, err := decoder.Int()
	if err != nil {
		return nil, err
	}
	globals := memory.NewGlobal()
	for i := 0; i < size; i++ {
		value, err := decoder.Any()
		if err != nil {
			return nil, fmt.Errorf("failed to decode global value at index %d: %w", i, err)
		}
		mutable, err := decoder.Bool()
		if err != nil {
			return nil, fmt.Errorf("failed to decode global mutable at index %d: %w", i, err)
		}
		globals.Push(value, mutable)
	}
	return globals, nil
}

func decodeMemories(decoder *Decoder) ([]iface.Memory, error) {
	count, err := decoder.Int()
	if err != nil {
		return nil, err
	}
	memories := make([]iface.Memory, count)
	for i := 0; i < count; i++ {
		data, err := decoder.Bytes()
		if err != nil {
			return nil, fmt.Errorf("failed to decode memory data at index %d: %w", i, err)
		}
		numPages, err := decoder.Int()
		if err != nil {
			return nil, fmt.Errorf("failed to decode memory numPages at index %d: %w", i, err)
		}
		maxPages, err := decoder.Int()
		if err != nil {
			return nil, fmt.Errorf("failed to decode memory maxPages at index %d: %w", i, err)
		}
		mem := memory.NewMemory(numPages, maxPages)
		mem.Store(0, data)
		memories[i] = mem
	}
	return memories, nil
}

func decodeTables(decoder *Decoder) ([]*memory.Table, error) {
	count, err := decoder.Int()
	if err != nil {
		return nil, err
	}
	tables := make([]*memory.Table, count)
	for i := 0; i < count; i++ {
		elementType, err := decoder.Byte()
		if err != nil {
			return nil, fmt.Errorf("failed to decode table elementType at index %d: %w", i, err)
		}
		initialSize, err := decoder.Int()
		if err != nil {
			return nil, fmt.Errorf("failed to decode table initialSize at index %d: %w", i, err)
		}
		maxSize, err := decoder.Int()
		if err != nil {
			return nil, fmt.Errorf("failed to decode table maxSize at index %d: %w", i, err)
		}
		numElements, err := decoder.Int()
		if err != nil {
			return nil, fmt.Errorf("failed to decode table numElements at index %d: %w", i, err)
		}
		elements := make([]int, numElements)
		for j := 0; j < numElements; j++ {
			elements[j], err = decoder.Int()
			if err != nil {
				return nil, fmt.Errorf("failed to decode table element at index %d/%d: %w", i, j, err)
			}
		}
		tables[i] = &memory.Table{
			ElementType: elementType,
			InitialSize: initialSize,
			MaxSize:     maxSize,
			Elements:    elements,
		}
	}
	return tables, nil
}

func decodeCallStack(decoder *Decoder, module *Module, instance *Instance, store *Store) error {
	count, err := decoder.Int()
	if err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		frame, err := decodeCallFrame(decoder, module, instance, store)
		if err != nil {
			return fmt.Errorf("failed to decode call frame at index %d: %w", i, err)
		}
		instance.callStack.Push(frame)
	}
	return nil
}

func decodeCallFrame(decoder *Decoder, module *Module, instance *Instance, store *Store) (*execution.CallFrame, error) {
	// Stack
	stackSize, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode stack size: %w", err)
	}
	stackItems := make([]any, stackSize)
	for i := 0; i < stackSize; i++ {
		stackItems[i], err = decoder.Any()
		if err != nil {
			return nil, fmt.Errorf("failed to decode stack item %d: %w", i, err)
		}
	}

	// Locals
	localsSize, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode locals size: %w", err)
	}
	localItems := make([]any, localsSize)
	for i := 0; i < localsSize; i++ {
		localItems[i], err = decoder.Any()
		if err != nil {
			return nil, fmt.Errorf("failed to decode local item %d: %w", i, err)
		}
	}

	// BlockStack
	blockStackSize, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode block stack size: %w", err)
	}
	blockStack := memory.NewStack[execution.BlockFrame]()
	for i := 0; i < blockStackSize; i++ {
		startPos, err := decoder.Int()
		if err != nil {
			return nil, fmt.Errorf("failed to decode block frame %d: %w", i, err)
		}
		blockStack.Push(execution.BlockFrame{StartPos: startPos})
	}

	// Body position and checkpoint
	bodyPos, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode body pos: %w", err)
	}
	bodyCheckpoint, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode body checkpoint: %w", err)
	}

	// FunctionIndex
	functionIndex, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode function index: %w", err)
	}

	// NumParams, NumResults
	numParams, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode num params: %w", err)
	}
	numResults, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode num results: %w", err)
	}

	// Params
	paramsSize, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode params size: %w", err)
	}
	params := make([]any, paramsSize)
	for i := 0; i < paramsSize; i++ {
		params[i], err = decoder.Any()
		if err != nil {
			return nil, fmt.Errorf("failed to decode param %d: %w", i, err)
		}
	}

	// FunctionCallRequest, Done, TailCall, Condition
	functionCallRequest, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode function call request: %w", err)
	}
	done, err := decoder.Bool()
	if err != nil {
		return nil, fmt.Errorf("failed to decode done: %w", err)
	}
	tailCall, err := decoder.Bool()
	if err != nil {
		return nil, fmt.Errorf("failed to decode tail call: %w", err)
	}
	condition, err := decoder.Bool()
	if err != nil {
		return nil, fmt.Errorf("failed to decode condition: %w", err)
	}

	frame := &execution.CallFrame{
		FunctionIndex: functionIndex,
		Context: execution.Context{
			Stack:  memory.NewStack[any](stackItems...),
			Locals: memory.NewStack[any](localItems...),

			NumParams:  numParams,
			NumResults: numResults,
			Params:     params,

			FunctionCallRequest: functionCallRequest,
			Done:                done,
			TailCall:            tailCall,

			Condition:  condition,
			BlockStack: blockStack,

			Globals:  store.Globals,
			Memories: store.Memories,
			Tables:   store.Tables,

			FuncSignatures: instance.funcSignatures,
			TypeSignatures: instance.typeSignatures,
		},
	}

	// Resolve the function reference
	if module.IsImport(functionIndex) {
		extFunc, err := instance.getImportedFunc(functionIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to get imported function at index %d: %w", functionIndex, err)
		}
		frame.Function = extFunc
	} else if module.IsFunction(functionIndex) {
		fn := module.FunctionAt(functionIndex)
		frame.Function = fn

		body := binary.NewIterator(fn.Body)
		body.Seek(bodyPos)
		body.SetCheckpointTo(bodyCheckpoint)
		frame.Context.Body = body

		frame.Context.Blocks = fn.Blocks
	}

	return frame, nil
}

func toStateStore(encoder *Encoder, store *Store) error {
	if err := encodeGlobalItems(encoder, store.Globals); err != nil {
		return err
	}
	encodeMemories(encoder, store.Memories)
	encodeTables(encoder, store.Tables)
	return nil
}

func encodeGlobalItems(encoder *Encoder, globals *memory.Global) error {
	size := globals.Size()
	encoder.Int(size)
	for i := 0; i < size; i++ {
		value, mutable, err := globals.Get(i)
		if err != nil {
			return fmt.Errorf("failed to get global at index %d: %v", i, err)
		}
		encoder.Any(value)
		encoder.Bool(mutable)
	}
	return nil
}

func encodeMemories(encoder *Encoder, memories []iface.Memory) {
	encoder.Int(len(memories))
	for _, mem := range memories {
		encoder.Bytes(mem.Data())
		encoder.Int(mem.NumPages())
		encoder.Int(mem.MaxPages())
	}
}

func encodeTables(encoder *Encoder, tables []*memory.Table) {
	encoder.Int(len(tables))
	for _, table := range tables {
		encoder.Byte(table.ElementType)
		encoder.Int(table.InitialSize)
		encoder.Int(table.MaxSize)
		encoder.Int(len(table.Elements))
		for _, elem := range table.Elements {
			encoder.Int(elem)
		}
	}
}

func encodeCallStack(encoder *Encoder, instance *Instance) error {
	size := instance.callStack.Size()
	encoder.Int(size)
	for i := 0; i < size; i++ {
		frame := instance.callStack.At(i)
		encodeCallFrame(encoder, frame)
	}
	return nil
}

func encodeCallFrame(encoder *Encoder, frame *execution.CallFrame) {
	if frame.Context.Stack == nil {
		encoder.Int(0)
	} else {
		encoder.Int(frame.Context.Stack.Size())
		for i := 0; i < frame.Context.Stack.Size(); i++ {
			encoder.Any(frame.Context.Stack.At(i))
		}
	}

	if frame.Context.Locals == nil {
		encoder.Int(0)
	} else {
		encoder.Int(frame.Context.Locals.Size())
		for i := 0; i < frame.Context.Locals.Size(); i++ {
			encoder.Any(frame.Context.Locals.At(i))
		}
	}

	if frame.Context.BlockStack == nil {
		encoder.Int(0)
	} else {
		encoder.Int(frame.Context.BlockStack.Size())
		for i := 0; i < frame.Context.BlockStack.Size(); i++ {
			blockFrame := frame.Context.BlockStack.At(i)
			encoder.Int(blockFrame.StartPos)
		}
	}

	var bodyPos, bodyCheckpoint int
	if frame.Context.Body != nil {
		bodyPos = frame.Context.Body.Position()
		bodyCheckpoint = frame.Context.Body.Checkpoint()
	}
	encoder.Int(bodyPos)
	encoder.Int(bodyCheckpoint)

	encoder.Int(frame.FunctionIndex)

	encoder.Int(frame.Context.NumParams)
	encoder.Int(frame.Context.NumResults)
	encoder.Int(len(frame.Context.Params))
	for _, param := range frame.Context.Params {
		encoder.Any(param)
	}

	encoder.Int(frame.Context.FunctionCallRequest)
	encoder.Bool(frame.Context.Done)
	encoder.Bool(frame.Context.TailCall)
	encoder.Bool(frame.Context.Condition)
}
