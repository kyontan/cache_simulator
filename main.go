package main

import (
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"
)

type Packet struct {
	Time               float64
	Len                uint64
	Proto              string
	SrcIP, DstIP       net.IP
	SrcPort, DstPort   uint64
	IcmpType, IcmpCode uint64
}

func (p *Packet) String() string {
	switch p.Proto {
	case "tcp", "udp":
		return fmt.Sprintf("{Time:%f Len:%v Proto:%v SrcIP:%v DstIP:%v SrcPort:%v DstPort:%v}", p.Time, p.Len, p.Proto, p.SrcIP, p.DstIP, p.SrcPort, p.DstPort)
	case "icmp":
		return fmt.Sprintf("{Time:%f Len:%v Proto:%v SrcIP:%v DstIP:%v IcmpType:%v IcmpCode:%v}", p.Time, p.Len, p.Proto, p.SrcIP, p.DstIP, p.IcmpType, p.IcmpCode)
	default:
		return fmt.Sprintf("{Time:%f Len:%v Proto:%v SrcIP:%v DstIP:%v SrcPort:%v DstPort:%v IcmpType:%v IcmpCode:%v}", p.Time, p.Len, p.Proto, p.SrcIP, p.DstIP, p.SrcPort, p.DstPort, p.IcmpType, p.IcmpCode)
	}
}

// type FiveTuple [5]uint64
type FiveTuple struct {
	Proto, SrcIP, DstIP, SrcPort, DstPort uint64
}

// {Proto, SrcIP, DstIP, SrcPort or 0, DstPort or 0}
func (p *Packet) FiveTuple() *FiveTuple {
	var proto64 uint64
	for i := 0; i < len(p.Proto) && i < 5; i++ {
		proto64 = proto64 << 8
		proto64 = proto64 | uint64(p.Proto[i])
	}
	srcIp64 := uint64(binary.LittleEndian.Uint32(p.SrcIP[len(p.SrcIP)-4:]))
	dstIp64 := uint64(binary.LittleEndian.Uint32(p.DstIP[len(p.DstIP)-4:]))

	switch p.Proto {
	case "tcp", "udp":
		return &FiveTuple{proto64, srcIp64, dstIp64, p.SrcPort, p.DstPort}
	// case "icmp":
	// 	return FiveTuple{p.Proto, p.SrcIP, p.DstIP, 0, 0}
	default:
		return nil
		// return FiveTuple{proto64, srcIp64, dstIp64, 0, 0}
	}
}

func (f FiveTuple) String() string {
	proto64 := f.Proto
	srcIp64 := f.SrcIP
	dstIp64 := f.DstIP
	SrcPort := f.SrcPort
	DstPort := f.DstPort
	var proto string
	for i := 0; i < 5 && proto64 != 0; i++ {
		c := proto64 & 0xff
		proto = string(c) + proto
		proto64 = proto64 >> 8
	}

	srcIp := make([]byte, 8)
	binary.LittleEndian.PutUint64(srcIp, uint64(srcIp64))
	dstIp := make([]byte, 8)
	binary.LittleEndian.PutUint64(dstIp, uint64(dstIp64))

	return fmt.Sprintf("FiveTuple{%v, %v, %v, %v, %v}", proto, net.IP(srcIp[0:4]), net.IP(dstIp[0:4]), SrcPort, DstPort)
}

func parseCSVRecord(record []string) (*Packet, error) {
	packet := new(Packet)
	var err error

	if len(record) != 7 {
		return nil, errors.New("Record must have 7 fields, but not")
	}

	packet.Time, err = strconv.ParseFloat(record[0], 64)
	if err != nil {
		return nil, err
	}
	packet.Len, err = strconv.ParseUint(record[1], 10, 32)
	if err != nil {
		return nil, err
	}

	packet.SrcIP = net.ParseIP(record[2])
	packet.DstIP = net.ParseIP(record[3])
	packet.Proto = record[4]

	switch packet.Proto {
	case "tcp", "udp":
		if packet.SrcPort, err = strconv.ParseUint(record[5], 10, 16); err != nil {
			return nil, err
		}
		if packet.DstPort, err = strconv.ParseUint(record[6], 10, 16); err != nil {
			return nil, err
		}
	case "icmp":
		if packet.IcmpType, err = strconv.ParseUint(record[5], 10, 16); err != nil {
			return nil, err
		}
		if packet.IcmpCode, err = strconv.ParseUint(record[6], 10, 16); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(fmt.Sprintf("unknown packet proto: %s", packet.Proto))
	}

	return packet, nil
}

type CacheSimulatorStat struct {
	Type      string
	Parameter string
	Processed int
	Hit       int
}

type CacheSimulator interface {
	IsCached(p *Packet) (cached bool, idx *int)
	Process(p *Packet) (hit bool)
	GetStat() (stat CacheSimulatorStat)
}

type SimpleLRUCache struct {
	Cache        []FiveTuple
	CacheAge     []int
	CacheRefered []int
	CacheSize    uint
	Stat         CacheSimulatorStat
}

func NewSimpleLRUCache(size uint) *SimpleLRUCache {
	return &SimpleLRUCache{
		Cache:        make([]FiveTuple, size),
		CacheAge:     make([]int, size),
		CacheRefered: make([]int, size),
		CacheSize:    size,
		Stat: CacheSimulatorStat{
			Type:      "LRU",
			Parameter: fmt.Sprintf("Size:%v", size),
			Processed: 0,
			Hit:       0,
		},
	}
}

func (lc *SimpleLRUCache) IsCached(p *Packet) (bool, *int) {
	return lc.IsCachedWithFiveTuple(*p.FiveTuple())
}

func (lc *SimpleLRUCache) IsCachedWithFiveTuple(f FiveTuple) (bool, *int) {
	for i, cacheEntry := range lc.Cache {
		if cacheEntry == f {
			return true, &i
		}
	}

	return false, nil
}

func (lc *SimpleLRUCache) ReplaceOldestEntry(p *Packet) (replacedIdx int) {
	return lc.ReplaceOldestEntryWithFiveTuple(*p.FiveTuple())
}

func (lc *SimpleLRUCache) ReplaceOldestEntryWithFiveTuple(f FiveTuple) (replacedIdx int) {
	oldestCacheAge := -1
	oldestCacheAgeIdx := -1
	for i, age := range lc.CacheAge {
		if oldestCacheAge < age {
			oldestCacheAge = age
			oldestCacheAgeIdx = i
		}
	}

	// fmt.Printf("Replace cache entry idx:%v, age:%v, refered:%v, entry:%v\n", oldestCacheAgeIdx, oldestCacheAge, lc.CacheRefered[oldestCacheAgeIdx], lc.Cache[oldestCacheAgeIdx])
	lc.Cache[oldestCacheAgeIdx] = f
	lc.CacheAge[oldestCacheAgeIdx] = 0
	lc.CacheRefered[oldestCacheAgeIdx] = 0

	return oldestCacheAgeIdx
}

func (lc *SimpleLRUCache) Process(p *Packet) bool {
	hit := false
	hitIdx := -1

	// find cache
	if cached, idx := lc.IsCached(p); cached {
		hit = cached
		hitIdx = *idx
		lc.CacheRefered[hitIdx] += 1
	}

	for i, _ := range lc.Cache {
		lc.CacheAge[i] += 1
	}

	if hit {
		lc.Stat.Hit += 1
		lc.CacheAge[hitIdx] = 0
	}

	// replace cache entry if not hit
	if !hit {
		lc.ReplaceOldestEntry(p)
	}

	// fmt.Println(lc.Cache)
	// fmt.Println(lc.CacheAge)

	lc.Stat.Processed += 1

	return hit
}

func (lc *SimpleLRUCache) GetStat() CacheSimulatorStat {
	return lc.Stat
}

type LRUCacheWithLookAhead struct {
	SimpleLRUCache
}

func NewLRUCacheWithLookAhead(size uint) *LRUCacheWithLookAhead {
	return &LRUCacheWithLookAhead{*NewSimpleLRUCache(size)}
}

func (lc *LRUCacheWithLookAhead) IsCached(p *Packet) (bool, *int) {
	return lc.SimpleLRUCache.IsCached(p)
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

func (lc *LRUCacheWithLookAhead) Process(p *Packet) bool {
	hit := lc.SimpleLRUCache.Process(p)

	if p.Proto == "tcp" {
		swapped := p.FiveTuple().SwapSrcAndDst()

		if cached, _ := lc.IsCachedWithFiveTuple(swapped); !cached {
			lc.ReplaceOldestEntryWithFiveTuple(swapped)
		}
	}

	return hit
}

func (lc *LRUCacheWithLookAhead) GetStat() CacheSimulatorStat {
	return lc.SimpleLRUCache.GetStat()
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

	cache := NewSimpleLRUCache(cacheSize)
	// cache := NewLRUCacheWithLookAhead(cacheSize)

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

		cache.Process(packet)
		// fmt.Printf("Process packet %v, hit: %v\n", packet, hit)
		// fmt.Printf("%+v\n", packet.FiveTuple())
		// fmt.Println(time, len, proto, srcIP, srcPort, dstIP, dstPort, icmpCode, icmpType)
		if cache.GetStat().Processed%10000 == 0 {
			fmt.Printf("%d packets processed, Cache hit: %d, Rate: %f\n", cache.GetStat().Processed, cache.GetStat().Hit, float64(cache.GetStat().Hit)/float64(cache.GetStat().Processed))
		}
	}

	fmt.Printf("Total %d packets processed, Cache hit: %d, Rate: %f\n", cache.GetStat().Processed, cache.GetStat().Hit, float64(cache.GetStat().Hit)/float64(cache.GetStat().Processed))
}
