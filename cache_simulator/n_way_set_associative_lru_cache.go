package cache_simulator

import (
	"bytes"
	"encoding/binary"

	"github.com/sigurn/crc8"
)

type NWaySetAssociativeLRUCache struct {
	Sets []FullAssociativeLRUCache // len(Sets) = Size / Way, each size == Way
	Way  uint
	Size uint
}

func fiveTupleToBigEndianByteArray(f *FiveTuple) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, *f)
	return buf.Bytes()
}

func (cache *NWaySetAssociativeLRUCache) IsCached(p *Packet, update bool) (bool, *int) {
	return cache.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (cache *NWaySetAssociativeLRUCache) setIdxFromFiveTuple(f *FiveTuple) uint {
	// TODO: only 8 bit is supported
	switch cache.Size / cache.Way {
	case 256:
		crc_table := crc8.MakeTable(crc8.CRC8)
		return uint(crc8.Checksum(fiveTupleToBigEndianByteArray(f), crc_table))
	default:
		panic("Not implemented!")
	}
}

func (cache *NWaySetAssociativeLRUCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	setIdx := cache.setIdxFromFiveTuple(f)
	return cache.Sets[setIdx].IsCachedWithFiveTuple(f, update) // TODO: return meaningful value
}

func (cache *NWaySetAssociativeLRUCache) Cache(p *Packet) {
	cache.CacheFiveTuple(p.FiveTuple())
}

func (cache *NWaySetAssociativeLRUCache) CacheFiveTuple(f *FiveTuple) {
	setIdx := cache.setIdxFromFiveTuple(f)
	cache.Sets[setIdx].CacheFiveTuple(f)
}

func (cache *NWaySetAssociativeLRUCache) Clear() {
	panic("Not implemented")
}
