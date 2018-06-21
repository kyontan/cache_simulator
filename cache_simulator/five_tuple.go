package cache_simulator

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/mervin0502/pcaparser"
)

// uint8 + uint32 x 2 + uint16 x2 = 104 byte
type FiveTuple struct {
	Proto            pcaparser.IPv4Protocol
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

func StrToIPv4Protocol(proto string) pcaparser.IPv4Protocol {
	switch proto {
	case "ICMP", "icmp":
		return pcaparser.IP_ICMPType
	case "IGMP", "igmp":
		return pcaparser.IP_IGMPType
	case "IP", "ip":
		return pcaparser.IP_IPType
	case "TCP", "tcp":
		return pcaparser.IP_TCPType
	case "EGP", "egp":
		return pcaparser.IP_EGPType
	case "IGP", "igp":
		return pcaparser.IP_IGPType
	case "UDP", "udp":
		return pcaparser.IP_UDPType
	case "RSVP", "rsvp":
		return pcaparser.IP_RSVPType
	case "GRE", "gre":
		return pcaparser.IP_GREType
	case "ESP", "esp":
		return pcaparser.IP_ESPType
	case "AH", "ah":
		return pcaparser.IP_AHType
	case "EIGRP", "eigrp":
		return pcaparser.IP_EIGRPType
	case "OSPF", "ospf":
		return pcaparser.IP_OSPFType
	case "IPIP", "ipip":
		return pcaparser.IP_IPIPType
	case "VRRP", "vrrp":
		return pcaparser.IP_VRRPType
	case "L2TP", "l2tp":
		return pcaparser.IP_L2TPType
	default:
		panic("Can't match any of the pcaparser.IPv4Protocol")
	}
}

// {Proto, SrcIP, DstIP, SrcPort or 0, DstPort or 0}
func (p *Packet) FiveTuple() *FiveTuple {
	var proto64 uint64
	for i := 0; i < len(p.Proto) && i < 5; i++ {
		proto64 = proto64 << 8
		proto64 = proto64 | uint64(p.Proto[i])
	}

	ipv4_proto := StrToIPv4Protocol(p.Proto)
	switch ipv4_proto {
	case pcaparser.IP_TCPType, pcaparser.IP_UDPType:
		return &FiveTuple{ipv4_proto, ipToUInt32(p.SrcIP), ipToUInt32(p.DstIP), p.SrcPort, p.DstPort}
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
