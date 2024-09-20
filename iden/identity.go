package iden

import (
	"code.dogecoin.org/gossip/codec"
	"code.dogecoin.org/gossip/dnet"
)

const IdenMsgMinSize = 4 + 1 + 1 + 2 + 2 + 2 + 1 + 1 + 2
const IconMaxSize = 1588 // 1584 + type

var TagIdentity = dnet.NewTag("Iden")

type IdentityMsg struct { // 190+1584+104 = 1878
	Time    dnet.DogeTime // [4] Current time when this message is signed (use to detect changes) (Doge Epoch)
	Name    string        // [1][30] display name
	Bio     string        // [1][120] short biography
	Lat     int16         // [2] WGS84 +/- 90 degrees, 60 seconds + 6ths (nearest 305m)
	Long    int16         // [2] WGS84 +/- 180 degrees, 60 seconds + 3rds (nearest 610m)
	Country string        // [2] ISO 3166-1 alpha-2 code (optional)
	City    string        // [1][30] city name (optional)
	Nodes   [][]byte      // [1][32]xN public keys of nodes claimed by this identity
	Icon    []byte        // [2][1584] 48x48 compressed (see dogeicon.go)
}

func (msg *IdentityMsg) IsValid() bool {
	for _, pub := range msg.Nodes {
		if len(pub) != 32 {
			return false
		}
	}
	return (len(msg.Name) <= 30 &&
		len(msg.Bio) <= 120 &&
		(len(msg.Country) == 2 || len(msg.Country) == 0) &&
		len(msg.City) <= 30 &&
		len(msg.Icon) == 1584)
}

func (msg *IdentityMsg) Encode() []byte {
	if len(msg.Name) > 30 {
		panic("Invalid identity: name longer than 30")
	}
	if len(msg.Bio) > 120 {
		panic("Invalid identity: bio longer than 120")
	}
	if len(msg.City) > 30 {
		panic("Invalid identity: city longer than 30")
	}
	if len(msg.Icon) != 1584 {
		panic("Invalid identity: icon size not 1584")
	}
	e := codec.Encode(10 + 31 + 121 + 31 + 1584)
	e.UInt32le(uint32(msg.Time))
	e.VarString(msg.Name)
	e.VarString(msg.Bio)
	e.UInt16le(uint16(msg.Lat))
	e.UInt16le(uint16(msg.Long))
	e.PadString(2, msg.Country)
	e.VarString(msg.City)
	e.VarUInt(uint64(len(msg.Nodes)))
	for _, pub := range msg.Nodes {
		if len(pub) != 32 {
			panic("Invalid public key: must be 32 bytes")
		}
		e.Bytes(pub)
	}
	if len(msg.Icon) > IconMaxSize {
		panic("icon too large")
	}
	e.UInt16le(uint16(len(msg.Icon)))
	e.Bytes(msg.Icon)
	return e.Result()
}

func DecodeIdentityMsg(payload []byte) (msg IdentityMsg) {
	d := codec.Decode(payload)
	msg.Time = dnet.DogeTime(d.UInt32le())
	msg.Name = d.VarString()
	msg.Bio = d.VarString()
	msg.Lat = int16(d.UInt16le())
	msg.Long = int16(d.UInt16le())
	msg.Country = d.PadString(2)
	msg.City = d.VarString()
	nodes := d.VarUInt()
	for i := uint64(0); i < nodes; i++ {
		msg.Nodes = append(msg.Nodes, d.Bytes(32))
	}
	icsize := d.UInt16le()
	if icsize > IconMaxSize {
		panic("icon too large")
	}
	msg.Icon = d.Bytes(int(icsize))
	return
}
