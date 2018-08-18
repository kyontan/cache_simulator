package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/yosuke-furukawa/json5/encoding/json5"

	"routing_simulator/cache_simulator"
)

func parseCSVRecord(record []string) (*cache_simulator.Packet, error) {
	packet := new(cache_simulator.Packet)
	var err error

	// 7-tuple: [time] [len] [srcIP] [dstIP] [proto] [srcPort] [dstPort]
	// 8-tuple: [time] [srcIP] [srcPort] [dstIP] [dstPort] [proto] 0x[type (hex)] [len]

	var recordTimeStr, recordPacketLenStr, recordProtoStr, recordSrcIPStr, recordSrcPortStr, recordDstIPStr, recordDstPortStr string

	switch len(record) {
	case 8:
		recordTimeStr = record[0]
		recordSrcIPStr = record[1]
		recordSrcPortStr = record[2]
		recordDstIPStr = record[3]
		recordDstPortStr = record[4]
		recordProtoStr = record[5]
		recordPacketLenStr = record[7]
	case 7:
		recordTimeStr = record[0]
		recordPacketLenStr = record[1]
		recordSrcIPStr = record[2]
		recordDstIPStr = record[3]
		recordProtoStr = record[4]
		recordSrcPortStr = record[5]
		recordDstPortStr = record[6]
	default:
		return nil, fmt.Errorf("Expected record have 7 or 8 fields, but not: %d", len(record))
	}

	packet.Time, err = strconv.ParseFloat(recordTimeStr, 64)
	if err != nil {
		return nil, err
	}
	packetLen, err := strconv.ParseUint(recordPacketLenStr, 10, 32)
	if err != nil {
		return nil, err
	}
	packet.Len = uint32(packetLen)

	packet.SrcIP = net.ParseIP(recordSrcIPStr)
	packet.DstIP = net.ParseIP(recordDstIPStr)
	packet.Proto = strings.ToLower(recordProtoStr)

	switch packet.Proto {
	case "tcp", "udp":
		srcPort, err := strconv.ParseUint(recordSrcPortStr, 10, 16)
		if err != nil {
			return nil, err
		}
		packet.SrcPort = uint16(srcPort)

		dstPort, err := strconv.ParseUint(recordDstPortStr, 10, 16)
		if err != nil {
			return nil, err
		}
		packet.DstPort = uint16(dstPort)
	// case "icmp":
	// 	icmpType, err := strconv.ParseUint(record[5], 10, 16)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	packet.IcmpType = uint16(icmpType)
	// 	icmpCode, err := strconv.ParseUint(record[6], 10, 16)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	packet.IcmpCode = uint16(icmpCode)
	default:
		return nil, fmt.Errorf("unknown packet proto: %s", packet.Proto)
	}

	return packet, nil
}

func getProperCSVReader(fp *os.File) *csv.Reader {
	newReader := func(fp *os.File, comma rune) *csv.Reader {
		fp.Seek(0, 0)
		reader := csv.NewReader(fp)
		reader.Comma = comma

		return reader
	}

	tryRead := func(reader *csv.Reader) (bool, error) {
		record, err := reader.Read()

		if err == io.EOF {
			return true, nil
		}

		if err != nil {
			return false, err
		}

		return len(record) != 1, nil
	}

	for _, comma := range []rune{',', '\t', ' '} {
		if ok, _ := tryRead(newReader(fp, comma)); ok {
			return newReader(fp, comma)
		}
	}

	return nil
}

func runSimpleCacheSimulatorWithCSV(fp *os.File, sim *cache_simulator.SimpleCacheSimulator, printInterval int) {
	reader := getProperCSVReader(fp)

	if reader == nil {
		panic("Can't read input as valid tsv/csv file")
	}

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
			fmt.Println("Error:", err)
			continue
			// panic(err)
		}

		if packet.FiveTuple() == nil {
			continue
		}

		sim.Process(packet)
		if sim.GetStat().Processed%printInterval == 0 {
			fmt.Printf("%v\n", sim.GetStatString())
		}
	}
}

func main() {

	if len(os.Args) != 2 && len(os.Args) != 3 {
		fmt.Printf("%s cacheparam [tsv]\n", os.Args[0])
		os.Exit(1)
	}

	simulaterDefinitionBytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	var simlatorDefinition interface{}
	err = json5.Unmarshal(simulaterDefinitionBytes, &simlatorDefinition)
	if err != nil {
		panic(err)
	}

	cacheSim, err := cache_simulator.BuildSimpleCacheSimulator(simlatorDefinition)

	if err != nil {
		panic(err)
	}

	var fpCSV *os.File

	if len(os.Args) == 2 {
		fpCSV = os.Stdin
	} else {
		var err error
		fpCSV, err = os.Open(os.Args[2])

		if err != nil {
			panic(err)
		}
		defer fpCSV.Close()
	}

	runSimpleCacheSimulatorWithCSV(fpCSV, &cacheSim, 1)

	fmt.Printf("%v\n", cacheSim.GetStatString())
}
