package cache

import (
	"fmt"

	"hash/crc32"
)

type NWaySetAssociativeLFUCache struct {
	Sets []FullAssociativeLFUCache // len(Sets) = Size / Way, each size == Way
	Way  uint
	Size uint
}

func (cache *NWaySetAssociativeLFUCache) StatString() string {
	return ""
}

func (cache *NWaySetAssociativeLFUCache) IsCached(p *Packet, update bool) (bool, *int) {
	return cache.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (cache *NWaySetAssociativeLFUCache) setIdxFromFiveTuple(f *FiveTuple) uint {
	maxSetIdx := cache.Size / cache.Way
	crc := crc32.ChecksumIEEE(fiveTupleToBigEndianByteArray(f))
	return uint(crc) % maxSetIdx
}

func (cache *NWaySetAssociativeLFUCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	setIdx := cache.setIdxFromFiveTuple(f)
	return cache.Sets[setIdx].IsCachedWithFiveTuple(f, update) // TODO: return meaningful value
}

func (cache *NWaySetAssociativeLFUCache) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	setIdx := cache.setIdxFromFiveTuple(f)
	return cache.Sets[setIdx].CacheFiveTuple(f)
}

func (cache *NWaySetAssociativeLFUCache) InvalidateFiveTuple(f *FiveTuple) {
	setIdx := cache.setIdxFromFiveTuple(f)
	cache.Sets[setIdx].InvalidateFiveTuple(f)
}

func (cache *NWaySetAssociativeLFUCache) Clear() {
	panic("Not implemented")
}

func (cache *NWaySetAssociativeLFUCache) Description() string {
	return "NWaySetAssociativeLFUCache"
}

func (cache *NWaySetAssociativeLFUCache) ParameterString() string {
	return fmt.Sprintf("{\"Type\": \"%s\", \"Way\": %d, \"Size\": %d}", cache.Description(), cache.Way, cache.Size)
}

func NewNWaySetAssociativeLFUCache(size, way uint) *NWaySetAssociativeLFUCache {
	if size%way != 0 {
		panic("Size must be multiplier of way")
	}

	sets_size := size / way
	sets := make([]FullAssociativeLFUCache, sets_size)

	for i := uint(0); i < sets_size; i++ {
		sets[i] = *NewFullAssociativeLFUCache(way)
	}

	return &NWaySetAssociativeLFUCache{
		Sets: sets,
		Way:  way,
		Size: size,
	}
}
