package cache_simulator

import (
	"fmt"
)

type FullAssociativeLRUCache struct {
	Entries []FiveTuple
	Age     []int
	Refered []int
	Size    uint
}

func (cache *FullAssociativeLRUCache) StatString() string {
	return ""
}

func (cache *FullAssociativeLRUCache) IsCached(p *Packet, update bool) (bool, *int) {
	return cache.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (cache *FullAssociativeLRUCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	hit := false
	var hitIdx *int

	for i, cacheEntry := range cache.Entries {
		if cacheEntry == *f {
			hit = true
			hitIdx = &i
			break
		}
	}

	if hit && update {
		for i, _ := range cache.Entries {
			cache.Age[i] += 1
		}

		cache.Refered[*hitIdx] += 1
		cache.Age[*hitIdx] = 0
	}

	return hit, hitIdx
}

func (cache *FullAssociativeLRUCache) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	for i, _ := range cache.Entries {
		cache.Age[i] += 1
	}

	oldestAge := -1
	oldestAgeIdx := -1
	for i, age := range cache.Age {
		if cache.Entries[i] == (FiveTuple{}) {
			oldestAgeIdx = i
			break
		}

		if oldestAge < age {
			oldestAge = age
			oldestAgeIdx = i
		}
	}

	fiveTupleToReplace := cache.Entries[oldestAgeIdx]

	// fmt.Printf("Replace cache entry idx:%v, age:%v, refered:%v, entry:%v\n", oldestAgeIdx, oldestAge, cache.Refered[oldestAgeIdx], cache.Cache[oldestAgeIdx])
	cache.Entries[oldestAgeIdx] = *f
	cache.Age[oldestAgeIdx] = 0
	cache.Refered[oldestAgeIdx] = 0

	if fiveTupleToReplace == (FiveTuple{}) {
		return []*FiveTuple{}
	}

	return []*FiveTuple{&fiveTupleToReplace}
}

func (cache *FullAssociativeLRUCache) InvalidateFiveTuple(f *FiveTuple) {
	var hitIdx *int

	for i, cacheEntry := range cache.Entries {
		if cacheEntry == *f {
			hitIdx = &i
			break
		}
	}

	if hitIdx == nil {
		panic("entry not cached")
	}

	cache.Entries[*hitIdx] = FiveTuple{}
	cache.Age[*hitIdx] = 0
	cache.Refered[*hitIdx] = 0
}

func (cache *FullAssociativeLRUCache) Clear() {
	panic("Not implemented")
}

func (cache *FullAssociativeLRUCache) Description() string {
	return "FullAssociativeLRUCache"
}

func (cache *FullAssociativeLRUCache) ParameterString() string {
	return fmt.Sprintf("Size: %d", cache.Size)
}
