package dnet

import "encoding/binary"

type Tag4CC uint32 // Big-Endian Four Character Code

func (t Tag4CC) String() string {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(t))
	return string(buf[:])
}

func (t Tag4CC) Bytes() []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(t))
	return buf[:]
}

func NewTag(tag string) Tag4CC {
	return Tag4CC(binary.BigEndian.Uint32([]byte(tag)))
}

// well-known channels
var ChannelNode = NewTag("Node")
var ChannelIdentity = NewTag("Iden")
var ChannelChat = NewTag("Chat")
var ChannelShibeShop = NewTag("Shib")
var ChannelB0rk = NewTag("B0rk")

// well-known services
var ServiceCore = NewTag("Core")
