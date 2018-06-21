package cache_simulator

type FullAssociativeLRUCache struct {
	Entries []FiveTuple
	Age     []int
	Refered []int
	Size    uint
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

func (cache *FullAssociativeLRUCache) Cache(p *Packet) {
	cache.CacheFiveTuple(p.FiveTuple())
}

func (cache *FullAssociativeLRUCache) CacheFiveTuple(f *FiveTuple) {
	oldestAge := -1
	oldestAgeIdx := -1
	for i, age := range cache.Age {
		if oldestAge < age {
			oldestAge = age
			oldestAgeIdx = i
		}
	}

	// fmt.Printf("Replace cache entry idx:%v, age:%v, refered:%v, entry:%v\n", oldestAgeIdx, oldestAge, cache.Refered[oldestAgeIdx], cache.Cache[oldestAgeIdx])
	cache.Entries[oldestAgeIdx] = *f
	cache.Age[oldestAgeIdx] = 0
	cache.Refered[oldestAgeIdx] = 0
}

func (cache *FullAssociativeLRUCache) Clear() {
	panic("Not implemented")
}
