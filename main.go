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

	"github.com/yosuke-furukawa/json5/encoding/json5"

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

func runSimpleCacheSimulatorWithCSV(fp *os.File, sim *cache_simulator.SimpleCacheSimulator, printInterval int) {
	reader := csv.NewReader(fp)
	reader.Comma = '\t'

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

		sim.Process(packet)
		// fmt.Printf("Process packet %v, hit: %v\n", packet, hit)
		// fmt.Printf("%+v\n", packet.FiveTuple())
		// fmt.Println(time, len, proto, srcIP, srcPort, dstIP, dstPort, icmpCode, icmpType)
		if sim.GetStat().Processed%printInterval == 0 {
			fmt.Printf("%v, %v\n", sim.GetStat(), sim.Cache.StatString())
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

	fmt.Println(cacheSim.Cache.Description())
	fmt.Println(cacheSim.Cache.ParameterString())

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

	fmt.Printf("%v, %v\n", cacheSim.GetStat(), cacheSim.Cache.StatString())
}
