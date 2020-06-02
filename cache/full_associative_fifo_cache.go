package cache

import (
	"container/list"
	"fmt"
)

type FullAssociativeFIFOCache struct {
	Entries map[FiveTuple]*list.Element
	Size    uint

	evictList *list.List
}

type fullAssociativeFIFOCacheEntry struct {
	Refered   int
	FiveTuple FiveTuple
}

func (cache *FullAssociativeFIFOCache) StatString() string {
	return ""
}

func (cache *FullAssociativeFIFOCache) AssertImmutableCondition() {
	if int(cache.Size) < len(cache.Entries) {
		panic(fmt.Sprintln("len(cache.Entries):", len(cache.Entries), ", expected: less than or equal to", cache.Size))
	}

	if cache.evictList.Len() != int(cache.Size) {
		panic(fmt.Sprintln("cache.evictList.Len():", cache.evictList.Len(), ", expected: ", cache.Size))
	}
}

func (cache *FullAssociativeFIFOCache) IsCached(p *Packet, update bool) (bool, *int) {
	return cache.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (cache *FullAssociativeFIFOCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	hitElem, hit := cache.Entries[*f]

	cache.AssertImmutableCondition()

	if hit && update {
		// cache.evictList.MoveToFront(hitElem)

		// update refered count
		hitEntry := hitElem.Value.(fullAssociativeFIFOCacheEntry)
		hitElem.Value = fullAssociativeFIFOCacheEntry{
			Refered:   hitEntry.Refered + 1,
			FiveTuple: hitEntry.FiveTuple,
		}
	}

	cache.AssertImmutableCondition()

	return hit, nil
}

func (cache *FullAssociativeFIFOCache) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	cache.AssertImmutableCondition()

	evictedFiveTuples := []*FiveTuple{}

	if hit, _ := cache.IsCachedWithFiveTuple(f, true); hit {
		return evictedFiveTuples
	}

	oldestElem := cache.evictList.Back()

	replacedEntry := cache.evictList.Remove(oldestElem).(fullAssociativeFIFOCacheEntry)
	delete(cache.Entries, replacedEntry.FiveTuple)

	newEntry := fullAssociativeFIFOCacheEntry{
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

func (cache *FullAssociativeFIFOCache) InvalidateFiveTuple(f *FiveTuple) {
	hitElem, hit := cache.Entries[*f]

	if !hit {
		panic("entry not cached")
	}

	cache.evictList.Remove(hitElem)
	delete(cache.Entries, *f)

	cache.evictList.PushBack(fullAssociativeFIFOCacheEntry{})

	cache.AssertImmutableCondition()
}

func (cache *FullAssociativeFIFOCache) Clear() {
	panic("Not implemented")
}

func (cache *FullAssociativeFIFOCache) Description() string {
	return "FullAssociativeFIFOCache"
}

func (cache *FullAssociativeFIFOCache) ParameterString() string {
	return fmt.Sprintf("{\"Type\": \"%s\", \"Size\": %d}", cache.Description(), cache.Size)
}

func NewFullAssociativeFIFOCache(size uint) *FullAssociativeFIFOCache {
	evictList := list.New()

	for i := 0; i < int(size); i++ {
		evictList.PushBack(fullAssociativeFIFOCacheEntry{})
	}

	return &FullAssociativeFIFOCache{
		Entries:   map[FiveTuple]*list.Element{},
		Size:      size,
		evictList: evictList,
	}
}
