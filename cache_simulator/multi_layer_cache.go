package cache_simulator

import (
	"fmt"
)

type CacheWritePolicy int
type CacheInclusionPolicy int

const (
	WriteThrough CacheWritePolicy = iota
	WriteBack
	// WriteAround
)

const (
	Inclusive CacheInclusionPolicy = iota
	Exclusive
)

type MultiLayerCache struct {
	CacheLayers            []Cache
	CacheWritePolicies     []CacheWritePolicy
	CacheInclusionPolicies []CacheInclusionPolicy
	CacheReferedByLayer    []uint
	CacheReplacedByLayer   []uint
	CacheHitByLayer        []uint
}

func (c *MultiLayerCache) StatString() string {
	return fmt.Sprintf("Refered: %v, Replaced: %v, Hit: %v", c.CacheReferedByLayer, c.CacheReplacedByLayer, c.CacheHitByLayer)
}

func (c *MultiLayerCache) IsCached(p *Packet, update bool) (bool, *int) {
	return c.IsCachedWithFiveTuple(p.FiveTuple(), update)
}

func (c *MultiLayerCache) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	for i, cache := range c.CacheLayers {
		if update {
			c.CacheReferedByLayer[i] += 1
		}
		if hit, hitIdx := cache.IsCachedWithFiveTuple(f, update); hit {
			if update {
				c.CacheHitByLayer[i] += 1
			}
			return hit, hitIdx
		}
	}

	return false, nil
}

func (c *MultiLayerCache) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	fiveTuplesToCache := []*FiveTuple{f}
	evictedFiveTuples := []*FiveTuple{f}

	for i, cache := range c.CacheLayers {
		fiveTuplesToCacheNextLayer := []*FiveTuple{}

		for _, f := range fiveTuplesToCache {
			evictedFiveTuples = cache.CacheFiveTuple(f)
			c.CacheReplacedByLayer[i] += uint(len(evictedFiveTuples))

			if (i + 1) == len(c.CacheLayers) {
				continue
			}

			switch c.CacheWritePolicies[i] {
			case WriteBack:
				fiveTuplesToCacheNextLayer = append(fiveTuplesToCacheNextLayer, evictedFiveTuples...)
			case WriteThrough:
			}
		}

		fiveTuplesToCache = fiveTuplesToCacheNextLayer
	}

	return evictedFiveTuples
}

func (c *MultiLayerCache) Clear() {
	for _, cache := range c.CacheLayers {
		cache.Clear()
	}
}
