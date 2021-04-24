package metric

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewWindow(t *testing.T) {
	opts := WindowOpts{Size: 3}
	window := NewWindow(opts)
	for i := 0; i < opts.Size; i++ {
		window.Append(i, 1.0)
		window.Append(i, 2.0)
	}
	assert.Equal(t, float64(1), window.Bucket(0).Points[0])
	assert.Equal(t, float64(2), window.Bucket(2).Points[1])
	t.Log(*window)
	window.window[0].Add(0, 10)
	window.Add(1, 100)
	assert.Equal(t, float64(11), window.Bucket(0).Points[0])
	assert.Equal(t, float64(101), window.Bucket(1).Points[0])

	t.Log(*window)
	iter := window.Iterator(0, 6)
	for iter.Next() {
		t.Logf("iterator: iterCount:%d, %+v", iter.iteratedCount, iter.Bucket())
	}

	//window.ResetWindow()
	window.ResetBuckets([]int{0, 2})
	for i := 0; i < opts.Size; i++ {
		t.Logf("bucket:%d, points:%+v", i, window.Bucket(i))
	}
}
