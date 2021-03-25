/**
    @author: xiangqian5
    @date: 2021/3/23
    @note:
**/
package metric

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestNewRollingPolicy(t *testing.T) {
	w := NewWindow(WindowOpts{Size: 10})
	p := NewRollingPolicy(w, RollingPolicyOpts{BucketDuration: 300 * time.Millisecond})
	t.Logf("window:%+v", p.window)
	t.Logf("size:%d", p.size)
	t.Logf("bucketDuration:%d", p.bucketDuration)
	t.Logf("lastAppendTime:%v", p.lastAppendTime)
	t.Logf("offset:%d", p.offset)
	for i, v := range p.window.window {
		t.Logf("bucket %d:%+v", i, v)
	}

	rand.Seed(time.Now().Unix())
	time.Sleep(400 * time.Millisecond)
	p.Add(1)
	time.Sleep(201 * time.Millisecond)
	p.Add(2)
	t.Log("================")
	t.Logf("window:%+v", p.window)
	t.Logf("size:%d", p.size)
	t.Logf("bucketDuration:%d", p.bucketDuration)
	t.Logf("lastAppendTime:%v", p.lastAppendTime)
	t.Logf("offset:%d", p.offset)
	for i, v := range p.window.window {
		t.Logf("bucket %d:%+v", i, v)
	}

}

func TestRollingPolicy2(t *testing.T) {
	table := []map[string][]int{
		{
			"timeSleep":       []int{294, 3200},
			"offsetAndPoints": []int{0, 1, 0, 1},
		},
		{
			"timeSleep":       []int{305, 3200, 6400},
			"offsetAndPoints": []int{1, 1, 1, 1, 1, 1},
		},
	}

	for _, hm := range table {
		var totalTs, lastOffset int
		offsetAndPoints := hm["offsetAndPoints"]
		timeSleep := hm["timeSleep"]
		w := NewWindow(WindowOpts{Size: 10})
		policy := NewRollingPolicy(w, RollingPolicyOpts{BucketDuration: 300 * time.Millisecond})
		for i, n := range timeSleep {
			totalTs += n
			time.Sleep(time.Duration(n) * time.Millisecond)
			policy.Add(1)
			offset, points := offsetAndPoints[2*i], offsetAndPoints[2*i+1]
			fmt.Println(policy.timespan(), points, n, policy.lastAppendTime, policy.window.window, policy.window.window[0])
			assert.Equal(t, float64(points), policy.window.window[offset].Points[0])
			lastOffset = offset
			fmt.Println(totalTs, lastOffset, policy)
			for i, v := range policy.window.window {
				t.Logf("bucket %d:%+v", i, v)
			}
		}
	}
}

func Test2(t *testing.T) {
	w := NewWindow(WindowOpts{Size: 10})
	p := NewRollingPolicy(w, RollingPolicyOpts{BucketDuration: 300 * time.Millisecond})
	time.Sleep(time.Millisecond * time.Duration(305))
	p.Add(1)
	t.Log("================")
	t.Logf("window:%+v", p.window)
	t.Logf("size:%d", p.size)
	t.Logf("bucketDuration:%d", p.bucketDuration)
	t.Logf("lastAppendTime:%v", p.lastAppendTime)
	t.Logf("offset:%d", p.offset)
	for i, v := range p.window.window {
		t.Logf("bucket %d:%+v", i, v)
	}
	time.Sleep(time.Millisecond * time.Duration(6400))
	p.Add(1)
	t.Log("================")
	t.Logf("window:%+v", p.window)
	t.Logf("size:%d", p.size)
	t.Logf("bucketDuration:%d", p.bucketDuration)
	t.Logf("lastAppendTime:%v", p.lastAppendTime)
	t.Logf("offset:%d", p.offset)
	for i, v := range p.window.window {
		t.Logf("bucket %d:%+v", i, v)
	}
}
