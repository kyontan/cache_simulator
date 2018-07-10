package cache_simulator

import (
	"fmt"

	"github.com/koron/go-dproxy"
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

func buildCache(p dproxy.Proxy) (Cache, error) {
	cache_type, err := p.M("Type").String()

	if err != nil {
		return nil, err
	}

	var cache Cache

	switch cache_type {
	case "CacheWithLookAhead":
		innerCache, err := buildCache(p.M("InnerCache"))
		if err != nil {
			return cache, err
		}

		cache = &CacheWithLookAhead{
			InnerCache: innerCache,
		}
	case "FullAssociativeLRUCache":
		size, err := p.M("Size").Int64()
		if err != nil {
			return cache, err
		}

		cache = NewFullAssociativeLRUCache(uint(size))
	case "NWaySetAssociativeLRUCache":
		size, err := p.M("Size").Int64()
		if err != nil {
			return cache, err
		}

		way, err := p.M("Way").Int64()
		if err != nil {
			return cache, err
		}

		cache = NewNWaySetAssociativeLRUCache(uint(size), uint(way))
	case "MultiLayerCache":
		cacheLayersPS := p.M("CacheLayers").ProxySet()
		cachePoliciesPS := p.M("CachePolicies").ProxySet()
		cacheLayersLen := cacheLayersPS.Len()
		cachePoliciesLen := cachePoliciesPS.Len()

		if cachePoliciesLen != (cacheLayersLen - 1) {
			return cache, fmt.Errorf("`CachePolicies` (%d items) must have `CacheLayers` length - 1 (%d) items", cachePoliciesLen, cacheLayersLen-1)
		}

		cacheLayers := make([]Cache, cacheLayersLen)
		for i := 0; i < cacheLayersLen; i++ {
			cacheLayer, err := buildCache(cacheLayersPS.A(i))
			if err != nil {
				return cache, err
			}
			cacheLayers[i] = cacheLayer
		}

		cachePolicies := make([]CachePolicy, cachePoliciesLen)
		for i := 0; i < cachePoliciesLen; i++ {
			cachePolicyStr, err := cachePoliciesPS.A(i).String()
			if err != nil {
				return cache, err
			}

			cachePolicies[i] = StringToCachePolicy(cachePolicyStr)
		}

		cache = &MultiLayerCache{
			CacheLayers:          cacheLayers,
			CachePolicies:        cachePolicies,
			CacheReferedByLayer:  make([]uint, cacheLayersLen),
			CacheReplacedByLayer: make([]uint, cacheLayersLen),
			CacheHitByLayer:      make([]uint, cacheLayersLen),
		}
	default:
		return nil, fmt.Errorf("Unsupported cache type: %s", cache_type)
	}

	return cache, nil
}

func BuildSimpleCacheSimulator(json interface{}) (SimpleCacheSimulator, error) {
	p := dproxy.New(json)

	simType, err := p.M("Type").String()

	if err != nil {
		return SimpleCacheSimulator{}, err
	}

	if simType != "SimpleCacheSimulator" {
		return SimpleCacheSimulator{}, fmt.Errorf("Unsupported simulator type: %s", simType)
	}

	cacheProxy := p.M("Cache")

	cache, err := buildCache(cacheProxy)

	if err != nil {
		return SimpleCacheSimulator{}, err
	}

	sim := SimpleCacheSimulator{
		Cache: cache,
		Stat: NewCacheSimulatorStat(
			cache.Description(),
			cache.ParameterString(),
		),
	}

	return sim, nil
}
