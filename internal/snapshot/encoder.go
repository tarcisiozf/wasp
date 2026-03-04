package snapshot

import (
	"fmt"
	"io"

	"github.com/tarcisiozf/wasp/internal/binary/leb"
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
