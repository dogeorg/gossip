package dnet

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

const MaxMsgSize = 0x1000080 // 16MB (block size is 1MB; 16x=16MB + 128 header 0x80)
const HeaderSize = 108

type Message struct { // 108 bytes fixed size header
	Chan      Tag4CC // [4] Channel Name [big-endian]
	Tag       Tag4CC // [4] Message Name [big-endian]
	Size      uint32 // [4] Size of the payload (excluding header)
	PubKey    []byte // [32]byte
	Signature []byte // [64]byte
	Payload   []byte // ... message payload
	RawHdr    []byte // attached raw header
}

// Encode a message by signing the payload.
func EncodeMessage(channel Tag4CC, tag Tag4CC, key KeyPair, payload []byte) []byte {
	if len(payload) > MaxMsgSize {
		panic("EncodeMessage: message too large: " + strconv.Itoa(len(payload)))
	}
	msg := make([]byte, HeaderSize+len(payload))
	binary.BigEndian.PutUint32(msg[0:4], uint32(channel))
	binary.BigEndian.PutUint32(msg[4:8], uint32(tag))
	binary.LittleEndian.PutUint32(msg[8:12], uint32(len(payload)))
	copy(msg[12:44], key.Pub) // PubKey (32 bytes)
	copy(msg[44:108], ed25519.Sign(key.Priv, payload))
	copy(msg[108:], payload)
	return msg
}

// Re-encode a message given the payload and signature.
func ReEncodeMessage(channel Tag4CC, tag Tag4CC, pubkey PubKey, sig []byte, payload []byte) []byte {
	if len(payload) > MaxMsgSize {
		panic("EncodeMessage: message too large: " + strconv.Itoa(len(payload)))
	}
	msg := make([]byte, HeaderSize+len(payload))
	binary.BigEndian.PutUint32(msg[0:4], uint32(channel))
	binary.BigEndian.PutUint32(msg[4:8], uint32(tag))
	binary.LittleEndian.PutUint32(msg[8:12], uint32(len(payload)))
	copy(msg[12:44], pubkey) // PubKey (32 bytes)
	copy(msg[44:108], sig)   // ed25519.Sign (64 bytes)
	copy(msg[108:], payload)
	return msg
}

func DecodeHeader(buf []byte) (msg Message) {
	msg.Chan = Tag4CC(binary.BigEndian.Uint32(buf[0:4]))
	msg.Tag = Tag4CC(binary.BigEndian.Uint32(buf[4:8]))
	msg.Size = binary.LittleEndian.Uint32(buf[8:12])
	msg.PubKey = buf[12:44]     // [32]byte
	msg.Signature = buf[44:108] // [64]byte
	msg.RawHdr = buf[0:108]     // for message forwarding
	return
}

func ReadMessage(reader io.Reader) (Message, error) {
	// Read the message header
	buf := [HeaderSize]byte{}
	n, err := io.ReadAtLeast(reader, buf[:], HeaderSize)
	if err != nil {
		return Message{}, fmt.Errorf("short header: received %d bytes: %v", n, err)
	}
	// Decode the header
	msg := DecodeHeader(buf[:])
	if msg.Size > MaxMsgSize {
		return Message{}, fmt.Errorf("message too large: [%s] size is %d bytes", msg.Tag, msg.Size)
	}
	// Read the message payload
	msg.Payload = make([]byte, msg.Size)
	n, err = io.ReadAtLeast(reader, msg.Payload, int(msg.Size))
	if err != nil {
		return Message{}, fmt.Errorf("short payload: [%s] received %d of %d bytes: %v", msg.Tag, n, msg.Size, err)
	}
	// Verify signature
	if !ed25519.Verify(msg.PubKey, msg.Payload, msg.Signature) {
		return Message{}, fmt.Errorf("incorrect signature: [%s] message", msg.Tag)
	}
	return msg, nil
}

// Message View (zero-copy)

type MessageView interface {
	Valid() bool
	ChanTag() (Tag4CC, Tag4CC)
	Size() uint
	PubKey() []byte    // 32 bytes
	Signature() []byte // 64 bytes
	Payload() []byte
}

type MessageViewImpl struct {
	msg []byte
}

func MsgView(msg []byte) MessageView {
	return MessageViewImpl{msg: msg}
}

func (m MessageViewImpl) Valid() bool {
	if len(m.msg) < HeaderSize {
		return false
	}
	sz := m.Size()
	if sz > MaxMsgSize {
		return false
	}
	if len(m.msg) != HeaderSize+int(sz) {
		return false
	}
	if !ed25519.Verify(m.PubKey(), m.msg[HeaderSize:], m.Signature()) {
		return false
	}
	return true
}

func (m MessageViewImpl) ChanTag() (Tag4CC, Tag4CC) {
	return Tag4CC(binary.BigEndian.Uint32(m.msg[0:4])), Tag4CC(binary.BigEndian.Uint32(m.msg[4:8]))
}

func (m MessageViewImpl) Size() uint {
	return uint(binary.LittleEndian.Uint32(m.msg[8:12]))
}

func (m MessageViewImpl) PubKey() []byte { // PubKey 32 bytes
	return m.msg[12:44] // [32]byte
}

func (m MessageViewImpl) Signature() []byte { // Signature 64 bytes
	return m.msg[44:108] // [64]byte
}

func (m MessageViewImpl) Payload() []byte {
	return m.msg[108:]
}
