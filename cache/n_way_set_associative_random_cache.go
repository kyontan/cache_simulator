package cache

import (
	"fmt"

	"hash/crc32"
)

type NWaySetAssociativeRandomCache struct {
	Sets []FullAssociativeRandomCache // len(Sets) = Size / Way, each size == Way
	Way  uint
	Size uint
}

// func fiveTupleToBigEndianByteArray(f *FiveTuple) []byte {
// 	var buf bytes.Buffer
// 	binary.Write(&buf, binary.BigEndian, *f)
// 	return buf.Bytes()
// }

func (cache *NWaySetAssociativeRandomCache) StatString() string {
	return ""
}

func (cache *NWaySetAssociativeRandomCache) IsCached(p *Packet, update bool) (bool, *int) {
	return cache.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (cache *NWaySetAssociativeRandomCache) setIdxFromFiveTuple(f *FiveTuple) uint {
	maxSetIdx := cache.Size / cache.Way
	crc := crc32.ChecksumIEEE(fiveTupleToBigEndianByteArray(f))
	return uint(crc) % maxSetIdx
}

func (cache *NWaySetAssociativeRandomCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	setIdx := cache.setIdxFromFiveTuple(f)
	return cache.Sets[setIdx].IsCachedWithFiveTuple(f, update) // TODO: return meaningful value
}

func (cache *NWaySetAssociativeRandomCache) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	setIdx := cache.setIdxFromFiveTuple(f)
	return cache.Sets[setIdx].CacheFiveTuple(f)
}

func (cache *NWaySetAssociativeRandomCache) InvalidateFiveTuple(f *FiveTuple) {
	setIdx := cache.setIdxFromFiveTuple(f)
	cache.Sets[setIdx].InvalidateFiveTuple(f)
}

func (cache *NWaySetAssociativeRandomCache) Clear() {
	panic("Not implemented")
}

func (cache *NWaySetAssociativeRandomCache) Description() string {
	return "NWaySetAssociativeRandomCache"
}

func (cache *NWaySetAssociativeRandomCache) ParameterString() string {
	return fmt.Sprintf("{\"Type\": \"%s\", \"Way\": %d, \"Size\": %d}", cache.Description(), cache.Way, cache.Size)
}

func NewNWaySetAssociativeRandomCache(size, way uint) *NWaySetAssociativeRandomCache {
	if size%way != 0 {
		panic("Size must be multiplier of way")
	}

	sets_size := size / way
	sets := make([]FullAssociativeRandomCache, sets_size)

	for i := uint(0); i < sets_size; i++ {
		sets[i] = *NewFullAssociativeRandomCache(way)
	}

	return &NWaySetAssociativeRandomCache{
		Sets: sets,
		Way:  way,
		Size: size,
	}
}
