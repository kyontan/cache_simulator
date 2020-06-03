package cache

import (
	"container/list"
	"fmt"
)

type FullAssociativeLFUCache struct {
	Entries map[FiveTuple]*list.Element
	Size    uint

	evictList *list.List
}

type fullAssociativeLFUCacheEntry struct {
	Refered   int
	FiveTuple FiveTuple
}

func (cache *FullAssociativeLFUCache) StatString() string {
	return ""
}

func (cache *FullAssociativeLFUCache) AssertImmutableCondition() {
	if int(cache.Size) < len(cache.Entries) {
		panic(fmt.Sprintln("len(cache.Entries):", len(cache.Entries), ", expected: less than or equal to", cache.Size))
	}

	if cache.evictList.Len() != int(cache.Size) {
		panic(fmt.Sprintln("cache.evictList.Len():", cache.evictList.Len(), ", expected: ", cache.Size))
	}
}

func (cache *FullAssociativeLFUCache) IsCached(p *Packet, update bool) (bool, *int) {
	return cache.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (cache *FullAssociativeLFUCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	hitElem, hit := cache.Entries[*f]

	cache.AssertImmutableCondition()

	if hit && update {
		// update refered count
		hitEntry := hitElem.Value.(fullAssociativeLFUCacheEntry)
		hitElem.Value = fullAssociativeLFUCacheEntry{
			Refered:   hitEntry.Refered + 1,
			FiveTuple: hitEntry.FiveTuple,
		}

		fRefered := hitElem.Value.(fullAssociativeLFUCacheEntry).Refered
		el := hitElem.Prev()

		// find elem that is more frequently used than hitElem
		for el != nil && el.Value.(fullAssociativeLFUCacheEntry).Refered <= fRefered {
			el = el.Prev()
		}

		if el == nil {
			cache.evictList.MoveToFront(hitElem)
		} else {
			cache.evictList.MoveAfter(hitElem, el)
		}
	}

	cache.AssertImmutableCondition()

	return hit, nil
}

func (cache *FullAssociativeLFUCache) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	cache.AssertImmutableCondition()

	evictedFiveTuples := []*FiveTuple{}

	if hit, _ := cache.IsCachedWithFiveTuple(f, true); hit {
		return evictedFiveTuples
	}

	lfuElem := cache.evictList.Back()

	replacedEntry := cache.evictList.Remove(lfuElem).(fullAssociativeLFUCacheEntry)
	delete(cache.Entries, replacedEntry.FiveTuple)

	newEntry := fullAssociativeLFUCacheEntry{
		FiveTuple: *f,
	}

	oldLFUel := cache.evictList.Back()

	newElem := cache.evictList.PushBack(newEntry)

	// move newElem to successor of an elem that is refered at least 1 time.
	for oldLFUel != nil && oldLFUel.Value.(fullAssociativeLFUCacheEntry).Refered == 0 {
		oldLFUel = oldLFUel.Prev()
	}

	if oldLFUel == nil {
		cache.evictList.MoveToFront(newElem)
	} else {
		cache.evictList.MoveAfter(newElem, oldLFUel)
	}

	cache.Entries[*f] = newElem

	cache.AssertImmutableCondition()

	if replacedEntry.FiveTuple == (FiveTuple{}) {
		return evictedFiveTuples
	}

	evictedFiveTuples = append(evictedFiveTuples, &replacedEntry.FiveTuple)

	return evictedFiveTuples
}

func (cache *FullAssociativeLFUCache) InvalidateFiveTuple(f *FiveTuple) {
	hitElem, hit := cache.Entries[*f]

	if !hit {
		panic("entry not cached")
	}

	cache.evictList.Remove(hitElem)
	delete(cache.Entries, *f)

	cache.evictList.PushBack(fullAssociativeLFUCacheEntry{})

	cache.AssertImmutableCondition()
}

func (cache *FullAssociativeLFUCache) Clear() {
	panic("Not implemented")
}

func (cache *FullAssociativeLFUCache) Description() string {
	return "FullAssociativeLFUCache"
}

func (cache *FullAssociativeLFUCache) ParameterString() string {
	return fmt.Sprintf("{\"Type\": \"%s\", \"Size\": %d}", cache.Description(), cache.Size)
}

func NewFullAssociativeLFUCache(size uint) *FullAssociativeLFUCache {
	evictList := list.New()

	for i := 0; i < int(size); i++ {
		evictList.PushBack(fullAssociativeLFUCacheEntry{})
	}

	return &FullAssociativeLFUCache{
		Entries:   map[FiveTuple]*list.Element{},
		Size:      size,
		evictList: evictList,
	}
}
