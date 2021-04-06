package trace

import (
	"github.com/zombie-k/kylin/library/libutil/hash"
	"math/rand"
	"sync/atomic"
	"time"
)

const (
	slotLength = 2048
)

var ignored = []string{"metrics", "ping"}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// sampler decides whether a new trace should be sampled or not.
type sampler interface {
	IsSampled(tarceID uint64, operationName string) (bool, float32)
	Close() error
}

type probabilitySampling struct {
	probability float32
	slot        [slotLength]int64
}

func (p *probabilitySampling) IsSampled(traceID uint64, operationName string) (bool, float32) {
	for _, ignore := range ignored {
		if operationName == ignore {
			return false, 0
		}
	}
	now := time.Now().Unix()
	idx := hash.JenkinsOneAtTimeHash(operationName) % slotLength
	old := atomic.LoadInt64(&p.slot[idx])
	if old != now {
		atomic.SwapInt64(&p.slot[idx], now)
		return true, 1
	}
	return rand.Float32() < p.probability, p.probability
}

func (p *probabilitySampling) Close() error {
	return nil
}

func newSampler(probability float32) sampler {
	if probability <= 0 || probability > 1 {
		panic("probability P ∈ (0, 1]")
	}
	return &probabilitySampling{probability: probability}
}
