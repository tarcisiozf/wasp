package serialization

import (
	"fmt"
	"io"
	"math"

	"github.com/tarcisiozf/wasp/internal/binary/leb"
)

type Writer interface {
	io.Writer
	io.ByteWriter
}

type Encoder struct {
	dest Writer
}

func (enc *Encoder) Int(value int) {
	enc.VarUint(uint64(value))
}

func (enc *Encoder) VarUint(value uint64) {
	leb.WriteUint(enc.dest, value)
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

func (enc *Encoder) Bytes(bytes []byte) {
	enc.Int(len(bytes))
	enc.dest.Write(bytes)
}

func (enc *Encoder) RawBytes(bytes []byte) {
	enc.dest.Write(bytes)
}

func (enc *Encoder) Byte(b byte) {
	enc.dest.Write([]byte{b})
}

func (enc *Encoder) Float64(f float64) {
	enc.VarUint(math.Float64bits(f))
}

func NewEncoder(dest Writer) *Encoder {
	return &Encoder{dest: dest}
}
