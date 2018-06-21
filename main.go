package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"

	"routing_simulator/cache_simulator"
)

func parseCSVRecord(record []string) (*cache_simulator.Packet, error) {
	packet := new(cache_simulator.Packet)
	var err error

	if len(record) != 7 {
		return nil, errors.New("Record must have 7 fields, but not")
	}

	packet.Time, err = strconv.ParseFloat(record[0], 64)
	if err != nil {
		return nil, err
	}
	plen, err := strconv.ParseUint(record[1], 10, 32)
	if err != nil {
		return nil, err
	}
	packet.Len = uint32(plen)

	packet.SrcIP = net.ParseIP(record[2])
	packet.DstIP = net.ParseIP(record[3])
	packet.Proto = record[4]

	switch packet.Proto {
	case "tcp", "udp":
		srcPort, err := strconv.ParseUint(record[5], 10, 16)
		if err != nil {
			return nil, err
		}
		packet.SrcPort = uint16(srcPort)

		dstPort, err := strconv.ParseUint(record[6], 10, 16)
		if err != nil {
			return nil, err
		}
		packet.DstPort = uint16(dstPort)
	case "icmp":
		icmpType, err := strconv.ParseUint(record[5], 10, 16)
		if err != nil {
			return nil, err
		}
		packet.IcmpType = uint16(icmpType)
		icmpCode, err := strconv.ParseUint(record[6], 10, 16)
		if err != nil {
			return nil, err
		}
		packet.IcmpCode = uint16(icmpCode)
	default:
		return nil, errors.New(fmt.Sprintf("unknown packet proto: %s", packet.Proto))
	}

	return packet, nil
}

func main() {
	var fp *os.File
	var err error

	if len(os.Args) < 2 {
		fmt.Printf("%s cacheSize [tsv]\n", os.Args[0])
		os.Exit(1)
	}

	var cacheSize uint
	if cacheSizeInt, err := strconv.ParseInt(os.Args[1], 10, 64); err != nil {
		fmt.Println("Can't parse", os.Args[1], "as integer cacheSize, aborting")
		os.Exit(1)
	} else {
		cacheSize = uint(cacheSizeInt)
	}

	if len(os.Args) != 3 {
		fp = os.Stdin
	} else {
		fp, err = os.Open(os.Args[2])

		if err != nil {
			panic(err)
		}
		defer fp.Close()
	}

	reader := csv.NewReader(fp)
	reader.Comma = '\t'

	// cacheSim := cache_simulator.NewFullAssociativeLRUCacheSimulator(cacheSize)
	// cacheSim := cache_simulator.NewFullAssociativeLRUCacheWithLookAheadSimulator(cacheSize)
	cacheSim := cache_simulator.NewNWaySetAssociativeLRUCacheSimulator(cacheSize, 4)

	for i := 0; ; i += 1 {
		record, err := reader.Read()

		if err != nil {
			if err == io.EOF {
				break
			}

			switch err.(type) {
			case *csv.ParseError:
				continue
			default:
				fmt.Println(reflect.TypeOf(err))
				continue
			}
		}

		packet, err := parseCSVRecord(record)
		if err != nil {
			panic(err)
		}

		if packet.FiveTuple() == nil {
			continue
		}

		cacheSim.Process(packet)
		// fmt.Printf("Process packet %v, hit: %v\n", packet, hit)
		// fmt.Printf("%+v\n", packet.FiveTuple())
		// fmt.Println(time, len, proto, srcIP, srcPort, dstIP, dstPort, icmpCode, icmpType)
		if cacheSim.GetStat().Processed%10000 == 0 {
			fmt.Println(cacheSim.GetStat())
		}
	}

	fmt.Println(cacheSim.GetStat())
}
