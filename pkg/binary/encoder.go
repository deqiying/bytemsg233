package binary

import (
	"encoding/binary"
	"io"
)

// Encoder writes binary data
type Encoder struct {
	w io.Writer
}

// NewEncoder creates a new encoder
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// WriteVarint writes a variable-length integer
func (e *Encoder) WriteVarint(value uint64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, value)
	_, err := e.w.Write(buf[:n])
	return err
}

// WriteZigzag writes a zigzag-encoded integer
func (e *Encoder) WriteZigzag(value int64) error {
	return e.WriteVarint(ZigzagEncode(value))
}

// WriteString writes a length-prefixed string
func (e *Encoder) WriteString(s string) error {
	data := []byte(s)
	if err := e.WriteVarint(uint64(len(data))); err != nil {
		return err
	}
	_, err := e.w.Write(data)
	return err
}

// WriteBytes writes length-prefixed bytes
func (e *Encoder) WriteBytes(data []byte) error {
	if err := e.WriteVarint(uint64(len(data))); err != nil {
		return err
	}
	_, err := e.w.Write(data)
	return err
}

// WriteFieldHeader writes a field header (tag + wire type)
func (e *Encoder) WriteFieldHeader(tag int, wireType int) error {
	return e.WriteVarint(uint64(tag<<3 | wireType))
}

// ZigzagEncode converts int64 to uint64 using zigzag encoding
func ZigzagEncode(value int64) uint64 {
	return uint64((value << 1) ^ (value >> 63))
}

// ZigzagDecode converts uint64 to int64 using zigzag decoding
func ZigzagDecode(value uint64) int64 {
	return int64((value >> 1) ^ -(value & 1))
}
