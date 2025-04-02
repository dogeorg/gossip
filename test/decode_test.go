package test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"code.dogecoin.org/gossip/codec"
)

func TestStreamBytes(t *testing.T) {
	stream := codec.Decode(hx2b("01020304"))
	result := stream.Bytes(4)
	if !bytes.Equal(result, hx2b("01020304")) {
		t.Errorf("Bytes: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("Bytes: stream not complete")
	}
}

func TestStreamRest(t *testing.T) {
	stream := codec.Decode(hx2b("01020304"))
	stream.UInt16le()
	result := stream.Rest()
	if !bytes.Equal(result, hx2b("0304")) {
		t.Errorf("Rest: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("Rest: stream not complete")
	}
}

func TestStreamBool(t *testing.T) {
	stream := codec.Decode(hx2b("01"))
	result := stream.Bool()
	if result != true {
		t.Errorf("Bool: wrong value: %v", result)
	}
	if !stream.Complete() {
		t.Errorf("Bool: stream not complete")
	}
}

func TestStreamUint8(t *testing.T) {
	stream := codec.Decode(hx2b("69"))
	result := stream.UInt8()
	if result != 0x69 {
		t.Errorf("Uint8: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("Uint8: stream not complete")
	}
}

func TestStreamUint16le(t *testing.T) {
	stream := codec.Decode(hx2b("0102"))
	result := stream.UInt16le()
	if result != 0x0201 {
		t.Errorf("Uint16le: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("Uint16le: stream not complete")
	}
}

func TestStreamUint16be(t *testing.T) {
	stream := codec.Decode(hx2b("0102"))
	result := stream.UInt16be()
	if result != 0x0102 {
		t.Errorf("Uint16be: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("Uint16be: stream not complete")
	}
}

func TestStreamUint32le(t *testing.T) {
	stream := codec.Decode(hx2b("01020304"))
	result := stream.UInt32le()
	if result != 0x04030201 {
		t.Errorf("Uint32le: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("Uint32le: stream not complete")
	}
}

func TestStreamUint32be(t *testing.T) {
	stream := codec.Decode(hx2b("01020304"))
	result := stream.UInt32be()
	if result != 0x01020304 {
		t.Errorf("Uint32be: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("Uint32be: stream not complete")
	}
}

func TestStreamUint64le(t *testing.T) {
	stream := codec.Decode(hx2b("0102030405060708"))
	result := stream.UInt64le()
	if result != 0x0807060504030201 {
		t.Errorf("Uint64le: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("Uint64le: stream not complete")
	}
}

func TestStreamInt64le(t *testing.T) {
	stream := codec.Decode(hx2b("0102030405060780"))
	result := stream.Int64le()
	if result != -9221395093405892095 {
		t.Errorf("Int64le: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("Int64le: stream not complete")
	}
}

func TestStreamVarUint1(t *testing.T) {
	// Test VarUint with 1 byte
	stream := codec.Decode(hx2b("01"))
	result := stream.VarUInt()
	if result != 0x01 {
		t.Errorf("VarUint: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("VarUint: stream not complete")
	}
}

func TestStreamVarUint2(t *testing.T) {
	// Test VarUint with 2 bytes
	stream := codec.Decode(hx2b("FD0102"))
	result := stream.VarUInt()
	if result != 0x0201 {
		t.Errorf("VarUint: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("VarUint: stream not complete")
	}
}

func TestStreamVarUint4(t *testing.T) {
	// Test VarUint with 4 bytes
	stream := codec.Decode(hx2b("FE01020304"))
	result := stream.VarUInt()
	if result != 0x04030201 {
		t.Errorf("VarUint: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("VarUint: stream not complete")
	}
}

func TestStreamVarUint8(t *testing.T) {
	// Test VarUint with 8 bytes
	stream := codec.Decode(hx2b("FF0102030405060708"))
	result := stream.VarUInt()
	if result != 0x0807060504030201 {
		t.Errorf("VarUint: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("VarUint: stream not complete")
	}
}

func TestStreamVarString(t *testing.T) {
	// Test VarString
	stream := codec.Decode(hx2b("0548656c6c6f"))
	result := stream.VarString()
	if result != "Hello" {
		t.Errorf("VarString: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("VarString: stream not complete")
	}
}

func TestStreamPadString(t *testing.T) {
	// Test PadString
	stream := codec.Decode(hx2b("48656c6c6f000000"))
	result := stream.PadString(8)
	if result != "Hello" {
		t.Errorf("PadString: wrong value: %x", result)
	}
	if !stream.Complete() {
		t.Errorf("PadString: stream not complete")
	}
}

func TestOverrunBytes(t *testing.T) {
	// Test overrun of bytes
	stream := codec.Decode(hx2b("01020304"))
	if stream.Bytes(5) != nil {
		t.Errorf("Bytes: should return nil for overrun")
	}
	if stream.Valid() {
		t.Errorf("Bytes: stream should not be valid for overrun")
	}
	if stream.Complete() {
		t.Errorf("Bytes: stream should not be complete for overrun")
	}
}

func TestOverrunUint16le(t *testing.T) {
	// Test overrun of Uint16le
	stream := codec.Decode(hx2b("01"))
	if stream.UInt16le() != 0 {
		t.Errorf("Uint16le: stream should return 0 for overrun")
	}
	if stream.Valid() {
		t.Errorf("Uint16le: stream should not be valid for overrun")
	}
	if stream.Complete() {
		t.Errorf("Uint16le: stream should not be complete for overrun")
	}
}

func TestOverrunUint32le(t *testing.T) {
	// Test overrun of Uint32le
	stream := codec.Decode(hx2b("010203"))
	if stream.UInt32le() != 0 {
		t.Errorf("Uint32le: stream should return 0 for overrun")
	}
	if stream.Valid() {
		t.Errorf("Uint32le: stream should not be valid for overrun")
	}
	if stream.Complete() {
		t.Errorf("Uint32le: stream should not be complete for overrun")
	}
}

func TestOverrunUint64le(t *testing.T) {
	// Test overrun of Uint64le
	stream := codec.Decode(hx2b("01020304050607"))
	if stream.UInt64le() != 0 {
		t.Errorf("Uint64le: stream should return 0 for overrun")
	}
	if stream.Valid() {
		t.Errorf("Uint64le: stream should not be valid for overrun")
	}
	if stream.Complete() {
		t.Errorf("Uint64le: stream should not be complete for overrun")
	}
}

func TestOverrunVarUint(t *testing.T) {
	// Test overrun of VarUint
	stream := codec.Decode([]byte{})
	if stream.VarUInt() != 0 {
		t.Errorf("VarUint: should return 0 for overrun")
	}
	if stream.Valid() {
		t.Errorf("VarUint: stream should not be valid for overrun")
	}
	if stream.Complete() {
		t.Errorf("VarUint: stream should not be complete for overrun")
	}
}

// Test Heplers

func hx2b(str string) (bytes []byte) {
	bytes, err := hex.DecodeString(str)
	if err != nil {
		panic("bad fixture: " + str)
	}
	return
}
