package simulator

import (
	"fmt"

	"github.com/koron/go-dproxy"
	"github.com/kyontan/cache_simulator/cache"
)

type SimpleCacheSimulator struct {
	cache.Cache
	Stat CacheSimulatorStat
}

func (sim *SimpleCacheSimulator) Process(p *cache.Packet) bool {
	// find cache
	cached := cache.AccessCache(sim.Cache, p)

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

func (sim *SimpleCacheSimulator) GetStatString() string {
	stat := sim.Stat.String()

	stat = stat[0 : len(stat)-1]

	statDetail := sim.Cache.StatString()

	if statDetail == "" {
		stat += ", \"StatDetail\": null}"
	} else {
		stat += ", \"StatDetail\": " + statDetail + "}"
	}

	return stat
}

func NewCacheSimulatorStat(description, parameter string) CacheSimulatorStat {
	return CacheSimulatorStat{
		Type:      description,
		Parameter: parameter,
		Processed: 0,
		Hit:       0,
	}
}

func buildCache(p dproxy.Proxy) (cache.Cache, error) {
	cache_type, err := p.M("Type").String()

	if err != nil {
		return nil, err
	}

	var c cache.Cache

	switch cache_type {
	case "CacheWithLookAhead":
		innerCache, err := buildCache(p.M("InnerCache"))
		if err != nil {
			return c, err
		}

		c = &cache.CacheWithLookAhead{
			InnerCache: innerCache,
		}
	case "FullAssociativeLRUCache":
		size, err := p.M("Size").Int64()
		if err != nil {
			return c, err
		}

		c = cache.NewFullAssociativeLRUCache(uint(size))
	case "NWaySetAssociativeLRUCache":
		size, err := p.M("Size").Int64()
		if err != nil {
			return c, err
		}

		way, err := p.M("Way").Int64()
		if err != nil {
			return c, err
		}

		c = cache.NewNWaySetAssociativeLRUCache(uint(size), uint(way))
	case "MultiLayerCache":
		cacheLayersPS := p.M("CacheLayers").ProxySet()
		cachePoliciesPS := p.M("CachePolicies").ProxySet()
		cacheLayersLen := cacheLayersPS.Len()
		cachePoliciesLen := cachePoliciesPS.Len()

		if cachePoliciesLen != (cacheLayersLen - 1) {
			return c, fmt.Errorf("`CachePolicies` (%d items) must have `CacheLayers` length - 1 (%d) items", cachePoliciesLen, cacheLayersLen-1)
		}

		cacheLayers := make([]cache.Cache, cacheLayersLen)
		for i := 0; i < cacheLayersLen; i++ {
			cacheLayer, err := buildCache(cacheLayersPS.A(i))
			if err != nil {
				return c, err
			}
			cacheLayers[i] = cacheLayer
		}

		cachePolicies := make([]cache.CachePolicy, cachePoliciesLen)
		for i := 0; i < cachePoliciesLen; i++ {
			cachePolicyStr, err := cachePoliciesPS.A(i).String()
			if err != nil {
				return c, err
			}

			cachePolicies[i] = cache.StringToCachePolicy(cachePolicyStr)
		}

		c = &cache.MultiLayerCache{
			CacheLayers:          cacheLayers,
			CachePolicies:        cachePolicies,
			CacheReferedByLayer:  make([]uint, cacheLayersLen),
			CacheReplacedByLayer: make([]uint, cacheLayersLen),
			CacheHitByLayer:      make([]uint, cacheLayersLen),
		}
	default:
		return nil, fmt.Errorf("Unsupported cache type: %s", cache_type)
	}

	return c, nil
}

func BuildSimpleCacheSimulator(json interface{}) (*SimpleCacheSimulator, error) {
	p := dproxy.New(json)

	simType, err := p.M("Type").String()

	if err != nil {
		return nil, err
	}

	if simType != "SimpleCacheSimulator" {
		return nil, fmt.Errorf("Unsupported simulator type: %s", simType)
	}

	cacheProxy := p.M("Cache")

	cache, err := buildCache(cacheProxy)

	if err != nil {
		return nil, err
	}

	sim := &SimpleCacheSimulator{
		Cache: cache,
		Stat: NewCacheSimulatorStat(
			cache.Description(),
			cache.ParameterString(),
		),
	}

	return sim, nil
}
