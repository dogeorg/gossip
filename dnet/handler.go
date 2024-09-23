package dnet

import "encoding/binary"

const BindMessageSize = 4 + 4 + 32

type BindMessage struct {
	Version uint32
	Chan    Tag4CC
	PubKey  [32]byte
}

func (msg BindMessage) Encode() []byte {
	buf := make([]byte, BindMessageSize)
	binary.LittleEndian.PutUint32(buf[0:4], msg.Version)
	binary.BigEndian.PutUint32(buf[4:8], uint32(msg.Chan))
	copy(buf[8:40], msg.PubKey[:])
	return buf
}

func DecodeBindMessage(payload []byte) (msg BindMessage, ok bool) {
	if len(payload) == BindMessageSize {
		msg.Version = binary.LittleEndian.Uint32(payload[0:4])
		msg.Chan = Tag4CC(binary.BigEndian.Uint32(payload[4:8]))
		copy(msg.PubKey[:], payload[8:40])
		return msg, true
	}
	return BindMessage{}, false
}
