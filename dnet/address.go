package dnet

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
)

// Address is an IP:Port combination.
type Address struct {
	Host net.IP
	Port uint16
}

func (a Address) String() string {
	return net.JoinHostPort(a.Host.String(), strconv.Itoa(int(a.Port)))
}

func (a Address) IsValid() bool {
	return a.Port != 0 && (len(a.Host) == 16 || len(a.Host) == 4)
}

func (a Address) ToBytes() []byte {
	buf := [18]byte{}
	copy(buf[0:16], a.Host)
	binary.BigEndian.PutUint16(buf[16:], a.Port)
	return buf[:]
}

func (a Address) Equal(other Address) bool {
	return a.Host.Equal(other.Host) && a.Port == other.Port
}

func ParseAddress(hostport string) (Address, error) {
	hosts, ports, err := net.SplitHostPort(hostport)
	if err != nil {
		return Address{}, err
	}
	host := net.ParseIP(hosts)
	if host == nil {
		return Address{}, errors.New("bad ip")
	}
	port, err := strconv.Atoi(ports)
	if err != nil {
		return Address{}, err
	}
	if port < 0 || port > 65535 {
		return Address{}, errors.New("range")
	}
	return Address{Host: host, Port: uint16(port)}, nil
}

func AddressFromBytes(addr []byte) (Address, error) {
	if len(addr) != 18 {
		return Address{}, errors.New("wrong address length")
	}
	return Address{
		Host: net.IP(addr[0:16]),
		Port: binary.BigEndian.Uint16(addr[16:]),
	}, nil
}
