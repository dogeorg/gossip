package codec

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Decoder is a helper for decoding data from a byte slice.
type Decoder struct {
	buf []byte // buffer to decode
	pos int    // current (unread) byte position
	len int    // length of the buffer
}

// Decode creates a new Decoder from a byte slice.
func Decode(b []byte) *Decoder {
	return &Decoder{buf: b, len: len(b)}
}

// Valid returns true if all reads have succeeded without reading past the end of the data.
func (d *Decoder) Valid() bool {
	return d.pos <= d.len
}

// Complete returns true if all the data has been read; implies Valid()
func (d *Decoder) Complete() bool {
	return d.pos == d.len
}

// Has checks if there are at least n bytes remaining in the buffer.
func (d *Decoder) Has(n int) bool {
	return d.pos+n <= d.len
}

// Bytes reads n bytes from the buffer and advances the position.
func (d *Decoder) Bytes(num int) []byte {
	pos := d.pos
	d.pos += num
	if pos+num <= d.len {
		return d.buf[pos : pos+num]
	}
	return nil
}

// Rest reads all remaining bytes in the buffer.
func (d *Decoder) Rest() []byte {
	p := d.pos
	d.pos = d.len
	return d.buf[p:]
}

// Bool reads a boolean (byte) value from the buffer.
func (d *Decoder) Bool() bool {
	pos := d.pos
	d.pos += 1
	if pos+1 <= d.len {
		return d.buf[pos] != 0
	}
	return false
}

// UInt8 reads an unsigned 8-bit integer from the buffer.
func (d *Decoder) UInt8() uint8 {
	pos := d.pos
	d.pos += 1
	if pos+1 <= d.len {
		return d.buf[pos]
	}
	return 0
}

// UInt16le reads an unsigned 16-bit integer from the buffer in little-endian order.
func (d *Decoder) UInt16le() uint16 {
	pos := d.pos
	d.pos += 2
	if pos+2 <= d.len {
		return binary.LittleEndian.Uint16(d.buf[pos : pos+2])
	}
	return 0
}

// UInt16be reads an unsigned 16-bit integer from the buffer in big-endian order.
func (d *Decoder) UInt16be() uint16 {
	pos := d.pos
	d.pos += 2
	if pos+2 <= d.len {
		return binary.BigEndian.Uint16(d.buf[pos : pos+2])
	}
	return 0
}

// UInt32le reads an unsigned 32-bit integer from the buffer in little-endian order.
func (d *Decoder) UInt32le() uint32 {
	pos := d.pos
	d.pos += 4
	if pos+4 <= d.len {
		return binary.LittleEndian.Uint32(d.buf[pos : pos+4])
	}
	return 0
}

// UInt32be reads an unsigned 32-bit integer from the buffer in big-endian order.
func (d *Decoder) UInt32be() uint32 {
	pos := d.pos
	d.pos += 4
	if pos+4 <= d.len {
		return binary.BigEndian.Uint32(d.buf[pos : pos+4])
	}
	return 0
}

// UInt64le reads an unsigned 64-bit integer from the buffer in little-endian order.
func (d *Decoder) UInt64le() uint64 {
	pos := d.pos
	d.pos += 8
	if pos+8 <= d.len {
		return binary.LittleEndian.Uint64(d.buf[pos : pos+8])
	}
	return 0
}

// Int64le reads a signed 64-bit integer from the buffer in little-endian order.
func (d *Decoder) Int64le() int64 {
	pos := d.pos
	d.pos += 8
	if pos+8 <= d.len {
		return int64(binary.LittleEndian.Uint64(d.buf[pos : pos+8]))
	}
	return 0
}

// VarUInt reads a variable-length unsigned integer from the buffer.
func (d *Decoder) VarUInt() uint64 {
	pos := d.pos
	d.pos += 1
	if pos+1 <= d.len {
		val := d.buf[pos]
		if val < 253 {
			return uint64(val)
		}
		if val == 253 {
			return uint64(d.UInt16le())
		}
		if val == 254 {
			return uint64(d.UInt32le())
		}
		return d.UInt64le()
	}
	return 0
}

// VarString reads a variable-length string from the buffer.
func (d *Decoder) VarString() string {
	len := d.VarUInt()
	if len > math.MaxInt {
		panic(fmt.Sprintf("decoded string length too long (greater than max-int): %v", len))
	}
	data := d.Bytes(int(len))
	return string(data)
}

// PadString reads a fixed-length padded string from the buffer, trimming trailing null bytes.
func (d *Decoder) PadString(size int) string {
	data := d.Bytes(size)
	if size > 0 {
		end := size - 1
		for data[end] == 0 && end > 0 {
			end--
		}
		return string(data[:end+1])
	}
	return ""
}
