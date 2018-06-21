package cache_simulator

type CacheWithLookAhead struct {
	InnerCache Cache
}

func (c *CacheWithLookAhead) IsCached(p *Packet, update bool) (bool, *int) {
	return c.InnerCache.IsCached(p, update)
}

func (c *CacheWithLookAhead) IsCachedWithFiveTuple(f *FiveTuple, update bool) (bool, *int) {
	return c.InnerCache.IsCachedWithFiveTuple(f, update)
}

func (c *CacheWithLookAhead) CacheFiveTuple(f *FiveTuple) {
	c.InnerCache.CacheFiveTuple(f)

	proto64 := f.Proto
	var proto string
	for i := 0; i < 5 && proto64 != 0; i++ {
		c := proto64 & 0xff
		proto = string(c) + proto
		proto64 = proto64 >> 8
	}

	if proto == "tcp" {
		swapped := (*f).SwapSrcAndDst()

		if cached, _ := c.InnerCache.IsCachedWithFiveTuple(&swapped, false); !cached {
			c.InnerCache.CacheFiveTuple(&swapped)
		}
	}
}

func (c *CacheWithLookAhead) Cache(p *Packet) {
	c.CacheFiveTuple(p.FiveTuple())
}

func (c *CacheWithLookAhead) Clear() {
	c.InnerCache.Clear()
}
