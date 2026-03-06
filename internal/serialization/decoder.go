package serialization

import (
	"fmt"
	"io"
	"unsafe"
)

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
	return dec.BytesN(length)
}

func (dec *Decoder) BytesN(length int) ([]byte, error) {
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

func (dec *Decoder) Float64() (float64, error) {
	v, err := dec.VarUint()
	if err != nil {
		return 0, err
	}
	return float64FromBits(v), nil
}

func float64FromBits(v uint64) float64 {
	return *(*float64)(unsafe.Pointer(&v))
}
