package cache_simulator

import (
	"github.com/mervin0502/pcaparser"
)

type CacheWithLookAhead struct {
	InnerCache Cache
}

func (c *CacheWithLookAhead) StatString() string {
	return ""
}

func (c *CacheWithLookAhead) IsCached(p *Packet, update bool) (bool, *int) {
	return c.InnerCache.IsCached(p, update)
}

func (c *CacheWithLookAhead) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	return c.InnerCache.IsCachedWithFiveTuple(f, update)
}

func (c *CacheWithLookAhead) CacheFiveTuple(f *FiveTuple) []*FiveTuple {
	evictedFiveTuples := c.InnerCache.CacheFiveTuple(f)

	if f.Proto == pcaparser.IP_TCPType {
		swapped := (*f).SwapSrcAndDst()

		if cached, _ := c.InnerCache.IsCachedWithFiveTuple(&swapped, false); !cached {
			replaced_by_lookahead := c.InnerCache.CacheFiveTuple(&swapped)
			evictedFiveTuples = append(evictedFiveTuples, replaced_by_lookahead...)
		}
	}

	return evictedFiveTuples
}

func (c *CacheWithLookAhead) InvalidateFiveTuple(f *FiveTuple) {
	c.InnerCache.InvalidateFiveTuple(f)
}

func (c *CacheWithLookAhead) Clear() {
	c.InnerCache.Clear()
}

func (c *CacheWithLookAhead) Description() string {
	return "CacheWithLookAhead[" + c.InnerCache.Description() + "]"
}

func (c *CacheWithLookAhead) ParameterString() string {
	return "[" + c.InnerCache.ParameterString() + "]"
}
