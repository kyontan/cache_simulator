package cache_simulator

import (
	"fmt"
)

type CachePolicy int

const (
	WriteThrough CachePolicy = iota
	WriteBackInclusive
	WriteBackExclusive
)


type MultiLayerCache struct {
	CacheLayers          []Cache
	CachePolicies        []CachePolicy
	CacheReferedByLayer  []uint
	CacheReplacedByLayer []uint
	CacheHitByLayer      []uint
}

func (c *MultiLayerCache) StatString() string {
	return fmt.Sprintf("Refered: %v, Replaced: %v, Hit: %v", c.CacheReferedByLayer, c.CacheReplacedByLayer, c.CacheHitByLayer)
}

func (c *MultiLayerCache) IsCached(p *Packet, update bool) (bool, *int) {
	return c.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (c *MultiLayerCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	hit := false
	var hitLayerIdx *int

	for i, cache := range c.CacheLayers {
		if update {
			c.CacheReferedByLayer[i] += 1
		}
		if hitLayer, _ := cache.IsCachedWithFiveTuple(f, update); hitLayer {
			if update {
				c.CacheHitByLayer[i] += 1
			}
			hit = true
			hitLayerIdx = &i
			break
		}
	}

	// cache miss at least L1 (layerIdx == 0)
	if update && hitLayerIdx != nil && *hitLayerIdx != 0 {
		// cache one layer upper, and then cache one more upper cache, ...
		// for i := *hitLayerIdx - 1; 0 <= i; i-- {
		// 	if c.CachePolicies[i] == WriteBackExclusive {
		// 		// invalidate under layer
		// 		c.CacheLayers[i+1].InvalidateFiveTuple(f)
		// 	}

		// 	// cache upper layer
		// 	c.CacheFiveTupleToLayer(f, i)
		// }

		// cache upper-most layer
		if c.CachePolicies[*hitLayerIdx-1] == WriteBackExclusive {
			// invalidate under layer
			c.CacheLayers[*hitLayerIdx].InvalidateFiveTuple(f)
		}

		c.CacheFiveTuple(f)
	}

	return hit, hitLayerIdx
}

func (c *MultiLayerCache) CacheFiveTupleToLayer(f *FiveTuple, layerIdx int) []*FiveTuple {
	fiveTuplesToCache := []*FiveTuple{f}
	evictedFiveTuples := []*FiveTuple{}

	for i, cache := range c.CacheLayers[layerIdx:] {
		fiveTuplesToCacheNextLayer := []*FiveTuple{}

		for _, f := range fiveTuplesToCache {
			evictedFiveTuples = cache.CacheFiveTuple(f)
			c.CacheReplacedByLayer[i] += uint(len(evictedFiveTuples))

			if i == (len(c.CacheLayers) - 1) {
				continue
			}

			switch c.CachePolicies[i] {
			case WriteBackExclusive, WriteBackInclusive:
				fiveTuplesToCacheNextLayer = append(fiveTuplesToCacheNextLayer, evictedFiveTuples...)
			case WriteThrough:
				fiveTuplesToCacheNextLayer = fiveTuplesToCache
			}
		}

		fiveTuplesToCache = fiveTuplesToCacheNextLayer
	}

	return evictedFiveTuples
}

func (c *MultiLayerCache) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	return c.CacheFiveTupleToLayer(f, 0)
}

func (c *MultiLayerCache) InvalidateFiveTuple(f *FiveTuple) {
	panic("not implemented")
}

func (c *MultiLayerCache) Clear() {
	for _, cache := range c.CacheLayers {
		cache.Clear()
	}
}
