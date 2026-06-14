package binary

import (
	"encoding/binary"
	"io"
)

// ByteReader combines io.Reader and io.ByteReader
type ByteReader interface {
	io.Reader
	io.ByteReader
}

// Decoder reads binary data
type Decoder struct {
	r ByteReader
}

// NewDecoder creates a new decoder
func NewDecoder(r ByteReader) *Decoder {
	return &Decoder{r: r}
}

// ReadVarint reads a variable-length integer
func (d *Decoder) ReadVarint() (uint64, error) {
	return binary.ReadUvarint(d.r)
}

// ReadZigzag reads a zigzag-encoded integer
func (d *Decoder) ReadZigzag() (int64, error) {
	v, err := d.ReadVarint()
	if err != nil {
		return 0, err
	}
	return ZigzagDecode(v), nil
}

// ReadString reads a length-prefixed string
func (d *Decoder) ReadString() (string, error) {
	length, err := d.ReadVarint()
	if err != nil {
		return "", err
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(d.r, buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

// ReadBytes reads length-prefixed bytes
func (d *Decoder) ReadBytes() ([]byte, error) {
	length, err := d.ReadVarint()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(d.r, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// ReadFieldHeader reads a field header (tag + wire type)
func (d *Decoder) ReadFieldHeader() (tag int, wireType int, err error) {
	v, err := d.ReadVarint()
	if err != nil {
		return 0, 0, err
	}
	tag = int(v >> 3)
	wireType = int(v & 0x7)
	return tag, wireType, nil
}
