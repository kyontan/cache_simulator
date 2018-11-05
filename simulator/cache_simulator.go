package simulator

import (
	"fmt"

	"github.com/kyontan/cache_simulator/cache"
)

type CacheSimulatorStat struct {
	Type      string
	Parameter string
	Processed int
	Hit       int
}

func (css CacheSimulatorStat) String() string {
	return fmt.Sprintf("{\"Type\": \"%s\", \"Parameter\": %s, \"Processed\": %v, \"Hit\": %v, \"HitRate\": %v}", css.Type, css.Parameter, css.Processed, css.Hit, float64(css.Hit)/float64(css.Processed))
}

type CacheSimulator interface {
	Process(p *cache.Packet) (hit bool)
	GetStat() (stat CacheSimulatorStat)
}
