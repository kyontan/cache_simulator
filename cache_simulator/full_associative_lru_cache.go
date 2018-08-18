package cache_simulator

import (
	"container/list"
	"fmt"
)

type FullAssociativeLRUCache struct {
	Entries map[FiveTuple]*list.Element
	Size    uint

	evictList *list.List
}

type entry struct {
	Refered   int
	FiveTuple FiveTuple
}

func (cache *FullAssociativeLRUCache) StatString() string {
	return ""
}

func (cache *FullAssociativeLRUCache) IsCached(p *Packet, update bool) (bool, *int) {
	return cache.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (cache *FullAssociativeLRUCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	hitElem, hit := cache.Entries[*f]

	if hit && update {
		cache.evictList.MoveToFront(hitElem)
		hitEntry := hitElem.Value.(entry)
		hitElem.Value = entry{
			Refered:   hitEntry.Refered + 1,
			FiveTuple: hitEntry.FiveTuple,
		}
	}

	return hit, nil
}

func (cache *FullAssociativeLRUCache) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	oldestElem := cache.evictList.Back()

	replacedEntry := cache.evictList.Remove(oldestElem).(entry)
	delete(cache.Entries, replacedEntry.FiveTuple)

	newEntry := entry{
		FiveTuple: *f,
	}

	newElem := cache.evictList.PushFront(newEntry)
	cache.Entries[*f] = newElem

	if replacedEntry.FiveTuple == (FiveTuple{}) {
		return []*FiveTuple{}
	}

	return []*FiveTuple{&replacedEntry.FiveTuple}
}

func (cache *FullAssociativeLRUCache) InvalidateFiveTuple(f *FiveTuple) {
	hitElem, hit := cache.Entries[*f]

	if !hit {
		panic("entry not cached")
	}

	cache.evictList.Remove(hitElem)
	delete(cache.Entries, *f)

	cache.evictList.PushBack(entry{})
}

func (cache *FullAssociativeLRUCache) Clear() {
	panic("Not implemented")
}

func (cache *FullAssociativeLRUCache) Description() string {
	return "FullAssociativeLRUCache"
}

func (cache *FullAssociativeLRUCache) ParameterString() string {
	return fmt.Sprintf("{\"Type\": \"%s\", \"Size\": %d}", cache.Description(), cache.Size)
}
