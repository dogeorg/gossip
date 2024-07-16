package spec

import (
	"bufio"
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

const MaxMsgSize = 0x1000080 // 16MB (block size is 1MB; 16x=16MB + 128 header 0x80)

type PrivKey = ed25519.PrivateKey // [32]privkey then [32]pubkey
type PubKey = ed25519.PublicKey   // [32]pubkey

type Message struct { // 108 bytes fixed size header
	Chan      Tag4CC // [4] Channel Name [big-endian]
	Tag       Tag4CC // [4] Message Name [big-endian]
	Size      uint32 // [4] Size of the payload (excluding header)
	PubKey    []byte // [32]byte
	Signature []byte // [64]byte
	Payload   []byte // ... message payload
}

func EncodeMessage(channel Tag4CC, tag Tag4CC, privkey PrivKey, payload []byte) []byte {
	if len(payload) > MaxMsgSize {
		panic("EncodeMessage: message too large: " + strconv.Itoa(len(payload)))
	}
	msg := make([]byte, 108+len(payload))
	binary.BigEndian.PutUint32(msg[0:4], uint32(channel))
	binary.BigEndian.PutUint32(msg[4:8], uint32(tag))
	binary.LittleEndian.PutUint32(msg[8:12], uint32(len(payload)))
	copy(msg[12:44], privkey[32:]) // PubKey (32 bytes)
	copy(msg[44:108], ed25519.Sign(privkey, payload))
	copy(msg[108:], payload)
	return msg
}

func DecodeMessage(buf []byte) (msg Message) {
	msg.Chan = Tag4CC(binary.BigEndian.Uint32(buf[0:4]))
	msg.Tag = Tag4CC(binary.BigEndian.Uint32(buf[4:8]))
	msg.Size = binary.LittleEndian.Uint32(buf[8:12])
	msg.PubKey = buf[12:44]     // [32]byte
	msg.Signature = buf[44:108] // [64]byte
	msg.Payload = buf[108:]
	return
}

func ReadMessage(reader *bufio.Reader) (Message, error) {
	// Read the message header
	buf := [108]byte{}
	n, err := io.ReadFull(reader, buf[:])
	if err != nil {
		return Message{}, fmt.Errorf("short header: received %d bytes: %v", n, err)
	}
	// Decode the header
	msg := DecodeMessage(buf[:])
	if msg.Size > MaxMsgSize {
		return Message{}, fmt.Errorf("message too large: [%s] size is %d bytes", msg.Tag, msg.Size)
	}
	// Read the message payload
	msg.Payload = make([]byte, msg.Size)
	n, err = io.ReadFull(reader, msg.Payload)
	if err != nil {
		return Message{}, fmt.Errorf("short payload: [%s] received %d of %d bytes: %v", msg.Tag, n, msg.Size, err)
	}
	// Verify signature
	if !ed25519.Verify(msg.PubKey, msg.Payload, msg.Signature) {
		return Message{}, fmt.Errorf("incorrect signature: [%s] message", msg.Tag)
	}
	return msg, nil
}
