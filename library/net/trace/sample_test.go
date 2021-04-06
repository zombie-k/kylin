package trace

import "testing"

func TestProbabilitySampling(t *testing.T) {
	sampler := newSampler(0.01)
	t.Run("test one operationName", func(t *testing.T) {
		sampled, probability := sampler.IsSampled(0, "test123")
		t.Logf("sampled:%v, probability:%f", sampled, probability)
	})
	t.Run("test probability", func(t *testing.T) {
		sampled, probability := sampler.IsSampled(0, "test_opt_2")
		t.Logf("sampled:%v, probability:%f", sampled, probability)
		count := 0
		for i := 0; i < 10000; i++ {
			sampled, prob := sampler.IsSampled(0, "test_opt_2")
			t.Logf("sampled:%v, probability:%f", sampled, prob)
			if sampled {
				count++
			}
		}
		t.Logf("count:%d", count)
	})
}
