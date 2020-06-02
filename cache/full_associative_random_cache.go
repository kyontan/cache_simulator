package cache

import (
	"container/list"
	"fmt"
	"math/rand"
	"time"
)

type FullAssociativeRandomCache struct {
	Entries map[FiveTuple]*list.Element
	Size    uint

	evictList *list.List
}

type fullAssociativeRandomCacheEntry struct {
	Refered   int
	FiveTuple FiveTuple
}

func init () {
	rand.Seed(time.Now().UnixNano())
}

func (cache *FullAssociativeRandomCache) StatString() string {
	return ""
}

func (cache *FullAssociativeRandomCache) AssertImmutableCondition() {
	if int(cache.Size) < len(cache.Entries) {
		panic(fmt.Sprintln("len(cache.Entries):", len(cache.Entries), ", expected: less than or equal to", cache.Size))
	}

	if int(cache.Size) < cache.evictList.Len() {
		panic(fmt.Sprintln("cache.evictList.Len():", cache.evictList.Len(), ", expected: ", cache.Size))
	}
}

func (cache *FullAssociativeRandomCache) IsCached(p *Packet, update bool) (bool, *int) {
	return cache.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (cache *FullAssociativeRandomCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	hitElem, hit := cache.Entries[*f]

	cache.AssertImmutableCondition()

	if hit && update {
		// update refered count
		hitEntry := hitElem.Value.(fullAssociativeRandomCacheEntry)
		hitElem.Value = fullAssociativeRandomCacheEntry{
			Refered:   hitEntry.Refered + 1,
			FiveTuple: hitEntry.FiveTuple,
		}
	}

	cache.AssertImmutableCondition()

	return hit, nil
}

func (cache *FullAssociativeRandomCache) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	cache.AssertImmutableCondition()

	evictedFiveTuples := []*FiveTuple{}

	if hit, _ := cache.IsCachedWithFiveTuple(f, true); hit {
		return evictedFiveTuples
	}

	if len(cache.Entries) == int(cache.Size) {
		// need to evict

		evictIdx := rand.Intn(int(cache.Size))

		el := cache.evictList.Front()

		for i := 0; i < evictIdx; i++ {
			el = el.Next()
		}

		randomElem := el

		replacedEntry := cache.evictList.Remove(randomElem).(fullAssociativeRandomCacheEntry)
		delete(cache.Entries, replacedEntry.FiveTuple)

		evictedFiveTuples = append(evictedFiveTuples, &replacedEntry.FiveTuple)
	}

	newEntry := fullAssociativeRandomCacheEntry{
		FiveTuple: *f,
	}

	newElem := cache.evictList.PushFront(newEntry)
	cache.Entries[*f] = newElem

	cache.AssertImmutableCondition()

	return evictedFiveTuples
}

func (cache *FullAssociativeRandomCache) InvalidateFiveTuple(f *FiveTuple) {
	hitElem, hit := cache.Entries[*f]

	if !hit {
		panic("entry not cached")
	}

	cache.evictList.Remove(hitElem)
	delete(cache.Entries, *f)

	cache.AssertImmutableCondition()
}

func (cache *FullAssociativeRandomCache) Clear() {
	panic("Not implemented")
}

func (cache *FullAssociativeRandomCache) Description() string {
	return "FullAssociativeRandomCache"
}

func (cache *FullAssociativeRandomCache) ParameterString() string {
	return fmt.Sprintf("{\"Type\": \"%s\", \"Size\": %d}", cache.Description(), cache.Size)
}

func NewFullAssociativeRandomCache(size uint) *FullAssociativeRandomCache {
	evictList := list.New()

	return &FullAssociativeRandomCache{
		Entries:   map[FiveTuple]*list.Element{},
		Size:      size,
		evictList: evictList,
	}
}
