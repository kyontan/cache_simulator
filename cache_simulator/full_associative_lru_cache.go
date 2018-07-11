package cache_simulator

import (
	"container/list"
	"fmt"
)

type FullAssociativeLRUCache struct {
	Entries []*list.Element
	Refered []int
	Size    uint

	evictList *list.List
}

type entry struct {
	Index     int
	FiveTuple FiveTuple
}

func (cache *FullAssociativeLRUCache) StatString() string {
	return ""
}

func (cache *FullAssociativeLRUCache) IsCached(p *Packet, update bool) (bool, *int) {
	return cache.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (cache *FullAssociativeLRUCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	hit := false
	var hitElem *list.Element
	var hitIdx *int

	for i, elem := range cache.Entries {
		cacheEntry := elem.Value.(entry).FiveTuple
		if cacheEntry == *f {
			hit = true
			hitElem = elem
			hitIdx = &i
			break
		}
	}

	if hit && update {
		cache.evictList.MoveToFront(hitElem)
		cache.Refered[*hitIdx] += 1
	}

	return hit, hitIdx
}

func (cache *FullAssociativeLRUCache) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	oldestElem := cache.evictList.Back()

	newEntry := entry{
		Index:     oldestElem.Value.(entry).Index,
		FiveTuple: *f,
	}

	replacedEntry := cache.evictList.Remove(oldestElem).(entry)
	newElem := cache.evictList.PushFront(newEntry)
	cache.Entries[newEntry.Index] = newElem

	cache.Refered[replacedEntry.Index] = 0

	if replacedEntry.FiveTuple == (FiveTuple{}) {
		return []*FiveTuple{}
	}

	return []*FiveTuple{&replacedEntry.FiveTuple}
}

func (cache *FullAssociativeLRUCache) InvalidateFiveTuple(f *FiveTuple) {
	var hitElem *list.Element

	for _, elem := range cache.Entries {
		cacheEntry := elem.Value.(entry).FiveTuple
		if cacheEntry == *f {
			hitElem = elem
			break
		}
	}

	if hitElem == nil {
		panic("entry not cached")
	}

	hitIdx := hitElem.Value.(entry).Index
	hitElem.Value = entry{
		Index:     hitIdx,
		FiveTuple: FiveTuple{},
	}
	cache.evictList.MoveToBack(hitElem)
	cache.Refered[hitIdx] = 0
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
