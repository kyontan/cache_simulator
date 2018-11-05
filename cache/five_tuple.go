package cache

import (
	"encoding/binary"
	"fmt"
	"net"
)

type IPProtocol uint8

const (
	IP_ICMP   IPProtocol = 1
	IP_TCP    IPProtocol = 6
	IP_UDP    IPProtocol = 17
	IP_ICMPv6 IPProtocol = 58
	IP_L2TP   IPProtocol = 115
)

// uint8 + uint32 x 2 + uint16 x2 = 104 byte
type FiveTuple struct {
	Proto            IPProtocol
	SrcIP, DstIP     uint32
	SrcPort, DstPort uint16
}

func ipToUInt32(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func uint32ToIP(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

func StrToIPProtocol(proto string) IPProtocol {
	switch proto {
	case "ICMP", "icmp":
		return IP_ICMP
	case "TCP", "tcp":
		return IP_TCP
	case "ICMPv6", "icmpv6":
		return IP_ICMPv6
	case "UDP", "udp":
		return IP_UDP
	case "L2TP", "l2tp":
		return IP_L2TP
	default:
		panic("Can't match any of the known protocols")
	}
}

// {Proto, SrcIP, DstIP, SrcPort or 0, DstPort or 0}
func (p *Packet) FiveTuple() *FiveTuple {
	var proto64 uint64
	for i := 0; i < len(p.Proto) && i < 5; i++ {
		proto64 = proto64 << 8
		proto64 = proto64 | uint64(p.Proto[i])
	}

	proto := StrToIPProtocol(p.Proto)
	switch proto {
	case IP_TCP, IP_UDP:
		return &FiveTuple{proto, ipToUInt32(p.SrcIP), ipToUInt32(p.DstIP), p.SrcPort, p.DstPort}
	// case "icmp":
	// 	return FiveTuple{p.Proto, p.SrcIP, p.DstIP, 0, 0}
	default:
		return nil
		// return FiveTuple{proto64, srcIp64, dstIp64, 0, 0}
	}
}

func (f FiveTuple) SwapSrcAndDst() FiveTuple {
	return FiveTuple{
		Proto:   f.Proto,
		SrcIP:   f.DstIP,
		DstIP:   f.SrcIP,
		SrcPort: f.DstPort,
		DstPort: f.SrcPort,
	}
}

func (f FiveTuple) String() string {
	return fmt.Sprintf("FiveTuple{%v, %v, %v, %v, %v}", f.Proto, uint32ToIP(f.SrcIP), uint32ToIP(f.DstIP), f.SrcPort, f.DstPort)
}
