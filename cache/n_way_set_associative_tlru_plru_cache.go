package cache

import (
	"fmt"

	"hash/crc32"
)

type NWaySetAssociativeTreePLRUCache struct {
	Sets []FullAssociativeTreePLRUCache // len(Sets) = Size / Way, each size == Way
	Way  uint
	Size uint
}

// func fiveTupleToBigEndianByteArray(f *FiveTuple) []byte {
// 	var buf bytes.Buffer
// 	binary.Write(&buf, binary.BigEndian, *f)
// 	return buf.Bytes()
// }

func (cache *NWaySetAssociativeTreePLRUCache) StatString() string {
	return ""
}

func (cache *NWaySetAssociativeTreePLRUCache) IsCached(p *Packet, update bool) (bool, *int) {
	return cache.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (cache *NWaySetAssociativeTreePLRUCache) setIdxFromFiveTuple(f *FiveTuple) uint {
	maxSetIdx := cache.Size / cache.Way
	crc := crc32.ChecksumIEEE(fiveTupleToBigEndianByteArray(f))
	return uint(crc) % maxSetIdx
}

func (cache *NWaySetAssociativeTreePLRUCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	setIdx := cache.setIdxFromFiveTuple(f)
	return cache.Sets[setIdx].IsCachedWithFiveTuple(f, update) // TODO: return meaningful value
}

func (cache *NWaySetAssociativeTreePLRUCache) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	setIdx := cache.setIdxFromFiveTuple(f)
	return cache.Sets[setIdx].CacheFiveTuple(f)
}

func (cache *NWaySetAssociativeTreePLRUCache) InvalidateFiveTuple(f *FiveTuple) {
	setIdx := cache.setIdxFromFiveTuple(f)
	cache.Sets[setIdx].InvalidateFiveTuple(f)
}

func (cache *NWaySetAssociativeTreePLRUCache) Clear() {
	panic("Not implemented")
}

func (cache *NWaySetAssociativeTreePLRUCache) Description() string {
	return "NWaySetAssociativeTreePLRUCache"
}

func (cache *NWaySetAssociativeTreePLRUCache) ParameterString() string {
	return fmt.Sprintf("{\"Type\": \"%s\", \"Way\": %d, \"Size\": %d}", cache.Description(), cache.Way, cache.Size)
}

func NewNWaySetAssociativeTreePLRUCache(size, way uint) *NWaySetAssociativeTreePLRUCache {
	if way != 8 && way != 16 {
		panic("Way must be 8 or 16")
	}

	if size%way != 0 {
		panic("Size must be multiplier of way")
	}

	sets_size := size / way
	sets := make([]FullAssociativeTreePLRUCache, sets_size)

	for i := uint(0); i < sets_size; i++ {
		sets[i] = *NewFullAssociativeTreePLRUCache(way)
	}

	return &NWaySetAssociativeTreePLRUCache{
		Sets: sets,
		Way:  way,
		Size: size,
	}
}
