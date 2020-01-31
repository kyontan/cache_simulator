package cache

import (
	"fmt"
)

type FullAssociativeTreePLRUCache struct {
	Entries map[FiveTuple]uint
	Size    uint

	// heap-like tree
	// evictTree[0] == LSB. 2nd-least significant bit is evictTree[1] if zero, evictTree[2] if one
	// evictTree[1] == ...

	// MSB     LSB
	// [1] [2] [4] - 0 (000), 1 (001)
	//         [5] - 2 (010), 3 (011)
	//     [3] [6] - 4 (100), 5 (101)
	//         [7] - 6 (110), 7 (111)

	// MSB     LSB
	// [1] [1] [1] - 0 (000), 1 (001)
	//         [0] - 2 (010), 3 (011)
	//     [1] [1] - 4 (100), 5 (101)
	//         [0] - 6 (110), 7 (111)

	// [1] [2] [4] [8]  - 0, 1
	//             [9]  - 2, 3
	//         [5] [10] - 4, 5
	//             [11] - 6, 7
	//     [3] [6] [12] - 8, 9
	//             [13] - 10, 11
	//         [7] [14] - 12, 13
	//             [15] - 14, 15

	evictTree    []bool
	entryFromIdx map[uint]*FiveTuple
}

func (cache *FullAssociativeTreePLRUCache) AssertImmutableCondition() {
	if len(cache.Entries) != len(cache.Entries) {
		panic(fmt.Sprintf("len(Entries) = %v should equal to len(Entries) = %v\n", len(cache.Entries), len(cache.entryFromIdx)))
	}

	if cache.Size < uint(len(cache.Entries)) {
		panic(fmt.Sprintf("len(Entries) = %v should be less than or equal to %v\n", len(cache.Entries), cache.Size))
	}

	if cache.Size < uint(len(cache.entryFromIdx)) {
		panic(fmt.Sprintf("len(entryFromIdx) = %v should be less than or equal to %v\n", len(cache.entryFromIdx), cache.Size))
	}
}

func (cache *FullAssociativeTreePLRUCache) StatString() string {
	return ""
}

func (cache *FullAssociativeTreePLRUCache) IsCached(p *Packet, update bool) (bool, *int) {
	return cache.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (cache *FullAssociativeTreePLRUCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	// log.Println("@@@ IsCachedWithFiveTuple")
	hitElemIdx, hit := cache.Entries[*f]

	if hit && update {
		// log.Printf("$$$ IsCached: update: %+v\n", f)
		// log.Printf("elemIdx: %v\n", hitElemIdx)
		treeIdx := uint(0)
		for mask := cache.Size >> 1; 0 < mask; mask >>= 1 {
			// (    (false) == 0 == 1) == true                     // => not flip
			// (    (false)    0 == 1) == false                    // =>     flip
			// (     (true)    1 == 1) == true                     // =>     flip
			// (     (true)    1 == 1) == false                    // => not flip
			cache.evictTree[treeIdx] = (hitElemIdx&mask == 0)
			// log.Printf("treeIdx: %v\n", treeIdx)

			if hitElemIdx&mask == 0 {
				treeIdx = (2*(treeIdx+1) + 0) - 1
			} else {
				treeIdx = (2*(treeIdx+1) + 1) - 1
			}
		}
		// log.Println("@@@@@@ IsCachedWithFiveTuple (after update)")
	}

	return hit, nil
}

func (cache *FullAssociativeTreePLRUCache) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	cache.AssertImmutableCondition()
	// log.Println("@@@ CacheFiveTuple")

	evictedFiveTuples := []*FiveTuple{}

	if hit, _ := cache.IsCachedWithFiveTuple(f, true); hit {
		return evictedFiveTuples
	}

	elemIdx := uint(0)
	treeIdx := 0
	for bit := uint(0); 1 < (cache.Size >> bit); bit += 1 {
		// (    (false) == 0 == 1) == true                     // => not flip
		// (    (false)    0 == 1) == false                    // =>     flip
		// (     (true)    1 == 1) == true                     // =>     flip
		// (     (true)    1 == 1) == false                    // => not flip
		if bit != 0 {
			elemIdx <<= 1
		}
		if cache.evictTree[treeIdx] {
			elemIdx |= 1
		}

		currentTreeIdx := treeIdx
		// log.Printf("treeIdx: %v\n", treeIdx)

		if !cache.evictTree[treeIdx] {
			treeIdx = (2*(treeIdx+1) + 0) - 1
		} else {
			treeIdx = (2*(treeIdx+1) + 1) - 1
		}

		cache.evictTree[currentTreeIdx] = !cache.evictTree[currentTreeIdx]
	}

	evictedFiveTuple, hit := cache.entryFromIdx[elemIdx]
	if hit {
		evictedFiveTuples = append(evictedFiveTuples, evictedFiveTuple)
		delete(cache.entryFromIdx, elemIdx)
		delete(cache.Entries, *evictedFiveTuple)
	}

	// newElem := cache.evictList.PushFront(newEntry)
	cache.Entries[*f] = elemIdx
	cache.entryFromIdx[elemIdx] = f

	cache.AssertImmutableCondition()

	// log.Printf("elemIdx: %v\n", elemIdx)
	// log.Println("@@@@@@ CacheFiveTuple (after update)")

	return evictedFiveTuples
}

func (cache *FullAssociativeTreePLRUCache) InvalidateFiveTuple(f *FiveTuple) {
	hitElemIdx, hit := cache.Entries[*f]

	// log.Printf("$$$ Invalidate: update: %+v (idx: %v)\n", f, hitElemIdx)

	if !hit {
		panic("entry not cached")
	}

	// mark hitElemIdx (idx of f as oldest)
	treeIdx := uint(0)
	for mask := cache.Size >> 1; 0 < mask; mask >>= 1 {
		// do the opposite of IsCached(update: true)
		cache.evictTree[treeIdx] = (hitElemIdx&mask != 0)
		// log.Printf("treeIdx: %v\n", treeIdx)

		if hitElemIdx&mask == 0 {
			treeIdx = (2*(treeIdx+1) + 0) - 1
		} else {
			treeIdx = (2*(treeIdx+1) + 1) - 1
		}
	}

	// log.Println("@@@@@@ Invalidate (after update)")

	delete(cache.entryFromIdx, hitElemIdx)
	delete(cache.Entries, *f)
}

func (cache *FullAssociativeTreePLRUCache) Clear() {
	panic("Not implemented")
}

func (cache *FullAssociativeTreePLRUCache) Description() string {
	return "FullAssociativeTreePLRUCache"
}

func (cache *FullAssociativeTreePLRUCache) ParameterString() string {
	return fmt.Sprintf("{\"Type\": \"%s\", \"Size\": %d}", cache.Description(), cache.Size)
}

func NewFullAssociativeTreePLRUCache(size uint) *FullAssociativeTreePLRUCache {
	if size != 8 && size != 16 {
		panic("FullAssociativeTreePLRUCache should have size == 8 or 16")
	}

	return &FullAssociativeTreePLRUCache{
		Entries:      map[FiveTuple]uint{},
		Size:         size,
		evictTree:    make([]bool, size-1, size-1),
		entryFromIdx: map[uint]*FiveTuple{},
	}
}
