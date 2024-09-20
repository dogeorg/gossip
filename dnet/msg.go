package dnet

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

	"github.com/dogeorg/doge"
)

const MaxMsgSize = 0x1000080 // 16MB (block size is 1MB; 16x=16MB + 128 header 0x80)
const HeaderSize = 108

type Message struct { // 108 bytes fixed size header
	Chan      Tag4CC // [4] Channel Name [big-endian]
	Tag       Tag4CC // [4] Message Name [big-endian]
	Size      uint32 // [4] Size of the payload (excluding header)
	PubKey    []byte // [32]byte Schnorr PubKey
	Signature []byte // [64]byte Schnorr Signature
	Payload   []byte // ... message payload
	RawHdr    []byte // attached raw header
}

// Encode a message by signing the payload.
func EncodeMessage(channel Tag4CC, tag Tag4CC, key KeyPair, payload []byte) []byte {
	if len(payload) > MaxMsgSize {
		panic("EncodeMessage: message too large: " + strconv.Itoa(len(payload)))
	}
	sig, err := doge.SignMessage(key.Priv, payload)
	if err != nil { // ErrInvalidPrivateKey, ErrPrivateKeyIsZero
		panic("invalid private key")
	}
	msg := make([]byte, HeaderSize+len(payload))
	binary.BigEndian.PutUint32(msg[0:4], uint32(channel))
	binary.BigEndian.PutUint32(msg[4:8], uint32(tag))
	binary.LittleEndian.PutUint32(msg[8:12], uint32(len(payload)))
	copy(msg[12:44], key.Pub[:]) // Schnorr PubKey (32 bytes)
	copy(msg[44:108], sig[:])    // Schnorr Signature (64 bytes)
	copy(msg[108:], payload)
	return msg
}

// Encode a message by signing the payload.
func EncodeMessageRaw(channel Tag4CC, tag Tag4CC, key KeyPair, payload []byte) RawMessage {
	hdr := EncodeHeaderAndSign(channel, tag, key, payload)
	return RawMessage{Header: hdr, Payload: payload}
}

// Encode a message header by signing the payload.
func EncodeHeaderAndSign(channel Tag4CC, tag Tag4CC, key KeyPair, payload []byte) []byte {
	if len(payload) > MaxMsgSize {
		panic("EncodeMessage: message too large: " + strconv.Itoa(len(payload)))
	}
	sig, err := doge.SignMessage(key.Priv, payload)
	if err != nil { // ErrInvalidPrivateKey, ErrPrivateKeyIsZero
		panic("invalid private key")
	}
	hdr := make([]byte, HeaderSize)
	binary.BigEndian.PutUint32(hdr[0:4], uint32(channel))
	binary.BigEndian.PutUint32(hdr[4:8], uint32(tag))
	binary.LittleEndian.PutUint32(hdr[8:12], uint32(len(payload)))
	copy(hdr[12:44], key.Pub[:]) // Schnorr PubKey (32 bytes)
	copy(hdr[44:108], sig[:])    // Schnorr Signature (64 bytes)
	return hdr
}

// Encode a message header using an existing signature.
func ReEncodeHeader(channel Tag4CC, tag Tag4CC, pubkey PubKey, sig []byte, payload []byte) []byte {
	if len(payload) > MaxMsgSize {
		panic("EncodeMessage: message too large: " + strconv.Itoa(len(payload)))
	}
	hdr := make([]byte, HeaderSize)
	binary.BigEndian.PutUint32(hdr[0:4], uint32(channel))
	binary.BigEndian.PutUint32(hdr[4:8], uint32(tag))
	binary.LittleEndian.PutUint32(hdr[8:12], uint32(len(payload)))
	copy(hdr[12:44], pubkey[:]) // PubKey (32 bytes)
	copy(hdr[44:108], sig)      // Signature (64 bytes)
	return hdr
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
	pub := (*[32]byte)(msg.PubKey)
	sig := (*[64]byte)(msg.Signature)
	if !doge.VerifyMessage(pub, msg.Payload, sig) {
		return Message{}, fmt.Errorf("incorrect signature: [%s] message", msg.Tag)
	}
	return msg, nil
}

// Message View (zero-copy)

type MessageView interface {
	Valid() bool
	ChanTag() (Tag4CC, Tag4CC)
	Size() uint
	PubKey() *[32]byte
	Signature() *[64]byte
	Header() []byte
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
	if !doge.VerifyMessage(m.PubKey(), m.msg[HeaderSize:], m.Signature()) {
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

func (m MessageViewImpl) PubKey() *[32]byte {
	return (*[32]byte)(m.msg[12:44])
}

func (m MessageViewImpl) Signature() *[64]byte {
	return (*[64]byte)(m.msg[44:108])
}

func (m MessageViewImpl) Header() []byte {
	return m.msg[0:108]
}

func (m MessageViewImpl) Payload() []byte {
	return m.msg[108:]
}

// RawMessage is a pre-formed message ready for sending.

type RawMessage struct {
	Header  []byte // encoded header
	Payload []byte // encoded payload
}

func (m RawMessage) Send(to io.Writer) error {
	_, err := to.Write(m.Header)
	if err != nil {
		return err
	}
	_, err = to.Write(m.Payload)
	if err != nil {
		return err
	}
	return nil
}

// Re-encode a message given the payload and signature.
func ReEncodeMessage(channel Tag4CC, tag Tag4CC, pubkey PubKey, sig []byte, payload []byte) RawMessage {
	hdr := ReEncodeHeader(channel, tag, pubkey, sig, payload)
	return RawMessage{Header: hdr, Payload: payload}
}
