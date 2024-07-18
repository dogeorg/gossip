package dnet

import "encoding/binary"

type BindMessage struct {
	Version uint32
	Chan    Tag4CC
}

func (msg BindMessage) Encode() []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[0:4], msg.Version)
	binary.BigEndian.PutUint32(buf[4:8], uint32(msg.Chan))
	return buf
}

func DecodeBindMessage(payload []byte) (msg BindMessage, ok bool) {
	if len(payload) == 8 {
		msg.Version = binary.LittleEndian.Uint32(payload[0:4])
		msg.Chan = Tag4CC(binary.BigEndian.Uint32(payload[4:8]))
		return msg, true
	}
	return BindMessage{}, false
}
