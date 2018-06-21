package cache_simulator

import (
	"fmt"
)

type SimpleCacheSimulator struct {
	Cache
	Stat CacheSimulatorStat
}

func (sim *SimpleCacheSimulator) Process(p *Packet) bool {
	// find cache
	cached := AccessCache(sim.Cache, p)

	if cached {
		sim.Stat.Hit += 1
	} else {
		// replace cache entry if not hit
		sim.Cache.Cache(p)
	}

	// fmt.Println(falc.Cache)
	// fmt.Println(falc.Age)

	sim.Stat.Processed += 1

	return cached
}

func (sim *SimpleCacheSimulator) GetStat() CacheSimulatorStat {
	return sim.Stat
}

func NewFullAssociativeLRUCacheSimulator(size uint) *SimpleCacheSimulator {
	return &SimpleCacheSimulator{
		Cache: &FullAssociativeLRUCache{
			Entries: make([]FiveTuple, size),
			Age:     make([]int, size),
			Refered: make([]int, size),
			Size:    size,
		},
		Stat: CacheSimulatorStat{
			Type:      "LRU",
			Parameter: fmt.Sprintf("Size:%v", size),
			Processed: 0,
			Hit:       0,
		},
	}
}

func NewFullAssociativeLRUCacheWithLookAheadSimulator(size uint) *SimpleCacheSimulator {
	return &SimpleCacheSimulator{
		Cache: &CacheWithLookAhead{
			InnerCache: &FullAssociativeLRUCache{
				Entries: make([]FiveTuple, size),
				Age:     make([]int, size),
				Refered: make([]int, size),
				Size:    size,
			},
		},
		Stat: CacheSimulatorStat{
			Type:      "LRU with Look-Ahead",
			Parameter: fmt.Sprintf("Size:%v", size),
			Processed: 0,
			Hit:       0,
		},
	}
}
