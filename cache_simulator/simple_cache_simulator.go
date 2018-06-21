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

func NewNWaySetAssociativeLRUCacheSimulator(size uint, way uint) *SimpleCacheSimulator {
	if size%way != 0 {
		panic("Size must be multiplier of way")
	}

	sets_size := size / way
	sets := make([]FullAssociativeLRUCache, sets_size)

	for i := uint(0); i < sets_size; i++ {
		sets[i] = FullAssociativeLRUCache{
			Entries: make([]FiveTuple, way),
			Age:     make([]int, way),
			Refered: make([]int, way),
			Size:    way,
		}
	}

	cache_sim := SimpleCacheSimulator{
		Cache: &NWaySetAssociativeLRUCache{
			Sets: sets,
			Way:  way,
			Size: size,
		},
		Stat: CacheSimulatorStat{
			Type:      "N Way Set Associative LRU",
			Parameter: fmt.Sprintf("Way: %v, Size:%v", way, size),
			Processed: 0,
			Hit:       0,
		},
	}

	return &cache_sim
}

func NewNWaySetAssociativeLRUCacheWithLookAheadSimulator(size uint, way uint) *SimpleCacheSimulator {
	if size%way != 0 {
		panic("Size must be multiplier of way")
	}

	sets_size := size / way
	sets := make([]FullAssociativeLRUCache, sets_size)

	for i := uint(0); i < sets_size; i++ {
		sets[i] = FullAssociativeLRUCache{
			Entries: make([]FiveTuple, way),
			Age:     make([]int, way),
			Refered: make([]int, way),
			Size:    way,
		}
	}

	cache_sim := SimpleCacheSimulator{
		Cache: &CacheWithLookAhead{
			InnerCache: &NWaySetAssociativeLRUCache{
				Sets: sets,
				Way:  way,
				Size: size,
			},
		},
		Stat: CacheSimulatorStat{
			Type:      "N Way Set Associative LRU with Look-Ahead",
			Parameter: fmt.Sprintf("Way: %v, Size:%v", way, size),
			Processed: 0,
			Hit:       0,
		},
	}

	return &cache_sim
}
