package breaker

import (
	"github.com/stretchr/testify/assert"
	"github.com/zombie-k/kylin/library/stat/metric"
	xtime "github.com/zombie-k/kylin/library/time"
	"math"
	"math/rand"
	"testing"
	"time"
)

func getSRE() Breaker {
	return NewGroup(&Config{
		K:       2,
		Window:  xtime.Duration(1 * time.Second),
		Bucket:  10,
		Request: 100,
	}).Get("")
}

func getSREBreaker() *sreBreaker {
	counterOpts := metric.RollingCounterOpts{
		Size:           10,
		BucketDuration: time.Millisecond * 100,
	}
	stat := metric.NewRollingCounter(counterOpts)
	return &sreBreaker{
		stat:    stat,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
		request: 100,
		k:       2,
		state:   StateClosed,
	}
}

func markSuccessWithDuration(b Breaker, count int, sleep time.Duration) {
	for i := 0; i < count; i++ {
		b.MarkSuccess()
		time.Sleep(sleep)
	}
}

func markFailedWithDuration(b Breaker, count int, sleep time.Duration) {
	for i := 0; i < count; i++ {
		b.MarkFailed()
		time.Sleep(sleep)
	}
}

func markSuccess(b Breaker, count int) {
	for i := 0; i < count; i++ {
		b.MarkSuccess()
	}
}

func markFailed(b Breaker, count int) {
	for i := 0; i < count; i++ {
		b.MarkFailed()
	}
}
func TestSRE(t *testing.T) {
	b := getSRE()
	markSuccess(b, 80)
	t.Logf("b:%+v", b)
	err := b.Allow()
	t.Logf("err:%v", err)
	markSuccess(b, 120)
	err = b.Allow()
	t.Logf("err:%v", err)

	b = getSRE()
	markSuccess(b, 100)
	t.Logf("b:%+v", b)
	err = b.Allow()
	t.Logf("err:%v", err)
	t.Logf("b:%+v", b)
	markFailed(b, 200)
	err = b.Allow()
	t.Logf("err:%v", err)
	t.Logf("b:%+v", b)
}

func TestSRESelfProtection(t *testing.T) {
	t.Run("total request < 100", func(t *testing.T) {
		b := getSRE()
		markFailed(b, 99)
		assert.Equal(t, nil, b.Allow())
	})

	t.Run("total request > 100, total < 2 * success", func(t *testing.T) {
		b := getSRE()
		size := rand.Intn(10000000)
		succ := int(math.Ceil(float64(size))) + 1
		markSuccess(b, succ)
		markFailed(b, size-succ)
		assert.Equal(t, nil, b.Allow())
	})
}

func TestSRESummary(t *testing.T) {
	var (
		b     *sreBreaker
		succ  float64
		total int64
	)
	sleep := 50 * time.Millisecond
	t.Run("succ == total", func(t *testing.T) {
		b = getSREBreaker()
		markSuccessWithDuration(b, 10, sleep)
		succ, total = b.summary()
		//t.Logf("succ:%f, total:%d", succ, total)
		assert.Equal(t, int64(10), int64(succ))
		assert.Equal(t, int64(10), total)
	})

	t.Run("fail == total", func(t *testing.T) {
		b = getSREBreaker()
		markFailedWithDuration(b, 10, sleep)
		succ, total = b.summary()
		//t.Logf("succ:%f, total:%d", succ, total)
		assert.Equal(t, int64(0), int64(succ))
		assert.Equal(t, int64(10), total)
	})

	t.Run("succ = 1/2 * total, tail = 1/2 * total", func(t *testing.T) {
		b = getSREBreaker()
		markFailedWithDuration(b, 5, sleep)
		markSuccessWithDuration(b, 5, sleep)
		succ, total = b.summary()
		//t.Logf("succ:%f, total:%d", succ, total)
		assert.Equal(t, int64(5), int64(succ))
		assert.Equal(t, int64(10), total)
	})

	t.Run("auto reset rolling counter", func(t *testing.T) {
		succ, total = b.summary()
		//t.Logf("succ:%f, total:%d", succ, total)
		time.Sleep(time.Second)
		succ, total = b.summary()
		//t.Logf("succ:%f, total:%d", succ, total)
		assert.Equal(t, int64(0), int64(succ))
		assert.Equal(t, int64(0), total)
	})
}

func TestDropRatio(t *testing.T) {
	const ratio = math.Pi / 10
	const total = 100000
	const epsilon = 0.05
	var count int
	b := getSREBreaker()
	for i := 0; i < total; i++ {
		if b.dropOnRatio(ratio) {
			count++
		}
	}

	rat := float64(count) / float64(total)
	t.Logf("ratio:%f,rat:%f,epsilon:%f", ratio, rat, epsilon)
	assert.InEpsilon(t, ratio, rat, epsilon)
}
