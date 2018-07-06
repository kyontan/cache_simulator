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
		sim.Cache.CacheFiveTuple(p.FiveTuple())
	}

	// fmt.Println(falc.Cache)
	// fmt.Println(falc.Age)

	sim.Stat.Processed += 1

	return cached
}

func (sim *SimpleCacheSimulator) GetStat() CacheSimulatorStat {
	return sim.Stat
}

func NewFullAssociativeLRUCache(size uint) *FullAssociativeLRUCache {
	return &FullAssociativeLRUCache{
		Entries: make([]FiveTuple, size),
		Age:     make([]int, size),
		Refered: make([]int, size),
		Size:    size,
	}
}

func NewNWaySetAssociativeLRUCache(size, way uint) *NWaySetAssociativeLRUCache {
	if size%way != 0 {
		panic("Size must be multiplier of way")
	}

	sets_size := size / way
	sets := make([]FullAssociativeLRUCache, sets_size)

	for i := uint(0); i < sets_size; i++ {
		sets[i] = *NewFullAssociativeLRUCache(way)
	}

	return &NWaySetAssociativeLRUCache{
		Sets: sets,
		Way:  way,
		Size: size,
	}
}

func NewCacheSimulatorStat(description, parameter string) CacheSimulatorStat {
	return CacheSimulatorStat{
		Type:      description,
		Parameter: parameter,
		Processed: 0,
		Hit:       0,
	}
}

func NewFullAssociativeLRUCacheSimulator(size uint) *SimpleCacheSimulator {
	return &SimpleCacheSimulator{
		Cache: NewFullAssociativeLRUCache(size),
		Stat: NewCacheSimulatorStat(
			"Full Associative (LRU)",
			fmt.Sprintf("Size:%v", size)),
	}
}

func NewFullAssociativeLRUCacheWithLookAheadSimulator(size uint) *SimpleCacheSimulator {
	return &SimpleCacheSimulator{
		Cache: &CacheWithLookAhead{
			InnerCache: NewFullAssociativeLRUCache(size),
		},
		Stat: NewCacheSimulatorStat(
			"Full Associative (LRU with Look-Ahead)",
			fmt.Sprintf("Size:%v", size)),
	}
}

func NewNWaySetAssociativeLRUCacheSimulator(size, way uint) *SimpleCacheSimulator {
	return &SimpleCacheSimulator{
		Cache: NewNWaySetAssociativeLRUCache(size, way),
		Stat: NewCacheSimulatorStat(
			"N Way Set Associative (LRU)",
			fmt.Sprintf("Way: %v, Size:%v", way, size)),
	}
}

func NewNWaySetAssociativeLRUCacheWithLookAheadSimulator(size, way uint) *SimpleCacheSimulator {
	return &SimpleCacheSimulator{
		Cache: &CacheWithLookAhead{
			InnerCache: NewNWaySetAssociativeLRUCache(size, way),
		},
		Stat: NewCacheSimulatorStat(
			"N Way Set Associative (LRU with Look-Ahead)",
			fmt.Sprintf("Way: %v, Size:%v", way, size)),
	}
}
