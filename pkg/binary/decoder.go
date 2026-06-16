package binary

import (
	"encoding/binary"
	"errors"
	"io"
	"unsafe"
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

var errVarintOverflow = errors.New("binary: varint overflows a 64-bit integer")

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

// ReadFixed32 reads a fixed-width 32-bit little-endian value.
func (d *Decoder) ReadFixed32() (uint32, error) {
	var buf [4]byte
	_, err := io.ReadFull(d.r, buf[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

// ReadFixed64 reads a fixed-width 64-bit little-endian value.
func (d *Decoder) ReadFixed64() (uint64, error) {
	var buf [8]byte
	_, err := io.ReadFull(d.r, buf[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf[:]), nil
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

// SliceDecoder reads binary data directly from a byte slice.
type SliceDecoder struct {
	data []byte
	pos  int
}

// NewSliceDecoder creates a decoder optimized for in-memory payloads.
func NewSliceDecoder(data []byte) *SliceDecoder {
	return &SliceDecoder{data: data}
}

// Reset points the decoder at a new byte slice.
func (d *SliceDecoder) Reset(data []byte) {
	d.data = data
	d.pos = 0
}

// EOF reports whether the decoder consumed the full slice.
func (d *SliceDecoder) EOF() bool {
	return d.pos >= len(d.data)
}

// Remaining returns the number of unread bytes.
func (d *SliceDecoder) Remaining() int {
	return len(d.data) - d.pos
}

// ReadVarint reads a variable-length integer without io.Reader overhead.
func (d *SliceDecoder) ReadVarint() (uint64, error) {
	if d.pos >= len(d.data) {
		return 0, io.ErrUnexpectedEOF
	}
	b := d.data[d.pos]
	if b < 0x80 {
		d.pos++
		return uint64(b), nil
	}
	return d.readVarintSlow()
}

// ReadZigzag reads a zigzag-encoded integer.
func (d *SliceDecoder) ReadZigzag() (int64, error) {
	value, err := d.ReadVarint()
	if err != nil {
		return 0, err
	}
	return ZigzagDecode(value), nil
}

// ReadString reads a length-prefixed string.
func (d *SliceDecoder) ReadString() (string, error) {
	if d.pos >= len(d.data) {
		return "", io.ErrUnexpectedEOF
	}
	b := d.data[d.pos]
	var length uint64
	if b < 0x80 {
		d.pos++
		length = uint64(b)
	} else {
		var err error
		length, err = d.readVarintSlow()
		if err != nil {
			return "", err
		}
	}
	if length > uint64(len(d.data)-d.pos) {
		return "", io.ErrUnexpectedEOF
	}
	start := d.pos
	d.pos += int(length)
	return string(d.data[start:d.pos]), nil
}

// ReadStringView reads a length-prefixed string without copying.
//
// The returned string aliases the decoder input. Use it only when the input
// byte slice will stay alive and immutable for at least as long as the string.
func (d *SliceDecoder) ReadStringView() (string, error) {
	value, err := d.ReadBytesView()
	if err != nil {
		return "", err
	}
	if len(value) == 0 {
		return "", nil
	}
	return unsafe.String(unsafe.SliceData(value), len(value)), nil
}

// ReadBytes reads length-prefixed bytes.
func (d *SliceDecoder) ReadBytes() ([]byte, error) {
	value, err := d.ReadBytesView()
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), value...), nil
}

// ReadBytesView reads length-prefixed bytes without copying.
func (d *SliceDecoder) ReadBytesView() ([]byte, error) {
	if d.pos >= len(d.data) {
		return nil, io.ErrUnexpectedEOF
	}
	b := d.data[d.pos]
	var length uint64
	if b < 0x80 {
		d.pos++
		length = uint64(b)
	} else {
		var err error
		length, err = d.readVarintSlow()
		if err != nil {
			return nil, err
		}
	}
	if length > uint64(len(d.data)-d.pos) {
		return nil, io.ErrUnexpectedEOF
	}
	start := d.pos
	d.pos += int(length)
	return d.data[start:d.pos], nil
}

// ReadFixed32 reads a fixed-width 32-bit little-endian value.
func (d *SliceDecoder) ReadFixed32() (uint32, error) {
	if len(d.data)-d.pos < 4 {
		return 0, io.ErrUnexpectedEOF
	}
	value := binary.LittleEndian.Uint32(d.data[d.pos:])
	d.pos += 4
	return value, nil
}

// ReadFixed64 reads a fixed-width 64-bit little-endian value.
func (d *SliceDecoder) ReadFixed64() (uint64, error) {
	if len(d.data)-d.pos < 8 {
		return 0, io.ErrUnexpectedEOF
	}
	value := binary.LittleEndian.Uint64(d.data[d.pos:])
	d.pos += 8
	return value, nil
}

// ReadFieldHeader reads a field header (tag + wire type).
func (d *SliceDecoder) ReadFieldHeader() (tag int, wireType int, err error) {
	if d.pos >= len(d.data) {
		return 0, 0, io.ErrUnexpectedEOF
	}
	b := d.data[d.pos]
	var value uint64
	if b < 0x80 {
		d.pos++
		value = uint64(b)
	} else {
		value, err = d.readVarintSlow()
		if err != nil {
			return 0, 0, err
		}
	}
	return int(value >> 3), int(value & 0x7), nil
}

func (d *SliceDecoder) readVarintSlow() (uint64, error) {
	var value uint64
	for shift := uint(0); shift < 64; shift += 7 {
		if d.pos >= len(d.data) {
			return 0, io.ErrUnexpectedEOF
		}
		b := d.data[d.pos]
		d.pos++
		if b < 0x80 {
			if shift == 63 && b > 1 {
				return 0, errVarintOverflow
			}
			return value | uint64(b)<<shift, nil
		}
		value |= uint64(b&0x7f) << shift
	}
	return 0, errVarintOverflow
}
