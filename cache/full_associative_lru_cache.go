package cache

import (
	"container/list"
	"fmt"
)

type FullAssociativeLRUCache struct {
	Entries map[FiveTuple]*list.Element
	Size    uint

	evictList *list.List
}

type fullAssociativeLRUCacheEntry struct {
	Refered   int
	FiveTuple FiveTuple
}

func (cache *FullAssociativeLRUCache) StatString() string {
	return ""
}

func (cache *FullAssociativeLRUCache) AssertImmutableCondition() {
	if int(cache.Size) < len(cache.Entries) {
		panic(fmt.Sprintln("len(cache.Entries):", len(cache.Entries), ", expected: less than or equal to", cache.Size))
	}

	if cache.evictList.Len() != int(cache.Size) {
		panic(fmt.Sprintln("cache.evictList.Len():", cache.evictList.Len(), ", expected: ", cache.Size))
	}
}

func (cache *FullAssociativeLRUCache) IsCached(p *Packet, update bool) (bool, *int) {
	return cache.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (cache *FullAssociativeLRUCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	hitElem, hit := cache.Entries[*f]

	cache.AssertImmutableCondition()

	if hit && update {
		cache.evictList.MoveToFront(hitElem)

		// update refered count
		hitEntry := hitElem.Value.(fullAssociativeLRUCacheEntry)
		hitElem.Value = fullAssociativeLRUCacheEntry{
			Refered:   hitEntry.Refered + 1,
			FiveTuple: hitEntry.FiveTuple,
		}
	}

	cache.AssertImmutableCondition()

	return hit, nil
}

func (cache *FullAssociativeLRUCache) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	cache.AssertImmutableCondition()

	evictedFiveTuples := []*FiveTuple{}

	if hit, _ := cache.IsCachedWithFiveTuple(f, true); hit {
		return evictedFiveTuples
	}

	oldestElem := cache.evictList.Back()

	replacedEntry := cache.evictList.Remove(oldestElem).(fullAssociativeLRUCacheEntry)
	delete(cache.Entries, replacedEntry.FiveTuple)

	newEntry := fullAssociativeLRUCacheEntry{
		FiveTuple: *f,
	}

	newElem := cache.evictList.PushFront(newEntry)
	cache.Entries[*f] = newElem

	cache.AssertImmutableCondition()

	if replacedEntry.FiveTuple == (FiveTuple{}) {
		return evictedFiveTuples
	}

	evictedFiveTuples = append(evictedFiveTuples, &replacedEntry.FiveTuple)

	return evictedFiveTuples
}

func (cache *FullAssociativeLRUCache) InvalidateFiveTuple(f *FiveTuple) {
	hitElem, hit := cache.Entries[*f]

	if !hit {
		panic("entry not cached")
	}

	cache.evictList.Remove(hitElem)
	delete(cache.Entries, *f)

	cache.evictList.PushBack(fullAssociativeLRUCacheEntry{})

	cache.AssertImmutableCondition()
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

func NewFullAssociativeLRUCache(size uint) *FullAssociativeLRUCache {
	evictList := list.New()

	for i := 0; i < int(size); i++ {
		evictList.PushBack(fullAssociativeLRUCacheEntry{})
	}

	return &FullAssociativeLRUCache{
		Entries:   map[FiveTuple]*list.Element{},
		Size:      size,
		evictList: evictList,
	}
}
