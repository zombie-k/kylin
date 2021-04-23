package pool

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	xtime "github.com/zombie-k/kylin/library/time"
	"io"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestT1(t *testing.T) {
	t1 := make(map[uint8]chan string)
	go func() {
		for i := 0; ; i++ {
			fmt.Println("go producer i=", i, " len_t1=", len(t1))
			if l := len(t1); l > 0 {
				var req chan string
				var reqKey uint8
				for reqKey, req = range t1 {
					break
				}
				delete(t1, reqKey)
				fmt.Println("go producer i=", i, " reqKey=", reqKey)
				req <- "loop_" + strconv.Itoa(int(reqKey)) + "_" + strconv.Itoa(i)
			}
			time.Sleep(time.Second)
		}
	}()

	var j uint8
	l := sync.Mutex{}
	for j = 250; ; j++ {
		l.Lock()
		req := make(chan string)
		reqKey := j
		t1[reqKey] = req
		l.Unlock()
		select {
		case ret, ok := <-req:
			fmt.Println("recv <-", ret, ok)
		}
	}
}

type closer struct {
}

func (c *closer) Close() error {
	return nil
}

type connection struct {
	c    io.Closer
	pool Pool
}

func (c *connection) HandleQuick() {
	//	time.Sleep(1 * time.Millisecond)
}

func (c *connection) HandleNormal() {
	time.Sleep(20 * time.Millisecond)
}

func (c *connection) HandleSlow() {
	time.Sleep(500 * time.Millisecond)
}

func (c *connection) Close() {
	c.pool.Put(context.Background(), c.c, false)
}

func TestSliceGetPut(t *testing.T) {
	config := &Config{
		Active:      20,
		Idle:        20,
		IdleTimeout: xtime.Duration(100 * time.Millisecond),
		//WaitTimeout: xtime.Duration(100 * time.Millisecond),
		//Wait:        false,
		Wait: true,
	}
	pool := NewSlice(config)
	pool.New = func(ctx context.Context) (i io.Closer, err error) {
		return &closer{}, nil
	}

	conn, err := pool.Get(context.TODO())
	assert.Nil(t, err)
	c1 := connection{pool: pool, c: conn}
	c1.HandleNormal()
	c1.HandleSlow()
	c1.Close()
	conn, err = pool.Get(context.TODO())
	c1 = connection{pool: pool, c: conn}
	c1.Close()
	c1.HandleSlow()

	for i := 0; i < 10; i++ {
		tmp := i
		go func(val int) {
			c, err := pool.Get(context.TODO())
			fmt.Println("goroutine", c, err)
		}(tmp)
	}
	time.Sleep(time.Second)
}

func TestSlicePut(t *testing.T) {
	type connID struct {
		io.Closer
		id int
	}
	var id = 0
	config := &Config{
		Active:      1,
		Idle:        1,
		IdleTimeout: xtime.Duration(1 * time.Second),
		Wait:        false,
	}
	pool := NewSlice(config)
	pool.New = func(ctx context.Context) (i io.Closer, err error) {
		id = id + 1
		return &connID{id: id, Closer: &closer{}}, nil
	}
	conn, err := pool.Get(context.TODO())
	assert.Nil(t, err)
	conn1 := conn.(*connID)
	t.Logf("conn1:%+v", conn1)
	conn1.id = 10
	pool.Put(context.TODO(), conn, true)
	conn, err = pool.Get(context.TODO())
	assert.Nil(t, err)
	conn2 := conn.(*connID)
	t.Logf("conn2:%+v", conn2)
	assert.NotEqual(t, conn1.id, conn2.id)
}

func TestSliceIdelTimeout(t *testing.T) {
	type connID struct {
		io.Closer
		id int
	}
	var id = 0
	config := &Config{
		Active:      1,
		Idle:        1,
		IdleTimeout: xtime.Duration(1 * time.Millisecond),
	}
	pool := NewSlice(config)
	pool.New = func(ctx context.Context) (i io.Closer, err error) {
		id = id + 1
		return &connID{
			Closer: &closer{},
			id:     id,
		}, nil
	}
	conn, err := pool.Get(context.TODO())
	assert.Nil(t, err)
	conn1 := conn.(*connID)
	assert.Equal(t, 1, conn1.id)
	assert.Equal(t, 0, len(pool.freeItem))
	assert.Equal(t, 1, pool.active)

	pool.Put(context.TODO(), conn, false)
	assert.Equal(t, 1, len(pool.freeItem))
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, 1, len(pool.freeItem))
	//time.Sleep(100 * time.Millisecond)
	//assert.Equal(t, 0, len(pool.freeItem))
	//t.Logf("3 pool:%s", pool)

	conn, err = pool.Get(context.TODO())
	assert.Nil(t, err)
	conn2 := conn.(*connID)
	assert.Equal(t, 0, len(pool.freeItem))
	assert.Equal(t, 1, pool.active)
	assert.Equal(t, 2, conn2.id)
}

func TestSliceContextTimeout(t *testing.T) {
	config := &Config{
		Active:      1,
		Idle:        1,
		IdleTimeout: xtime.Duration(90 * time.Second),
		WaitTimeout: xtime.Duration(10 * time.Millisecond),
	}
	pool := NewSlice(config)
	pool.New = func(ctx context.Context) (i io.Closer, err error) {
		return &closer{}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	conn, err := pool.Get(ctx)
	assert.Nil(t, err)
	_, err = pool.Get(ctx)
	assert.NotNil(t, err)
	t.Logf("err:%s", err)
	pool.Put(context.TODO(), conn, false)
	_, err = pool.Get(ctx)
	assert.Nil(t, err)
	t.Logf("err:%s", err)
}

func TestSlicePoolExhausted(t *testing.T) {
	// test pool exhausted
	config := &Config{
		Active:      1,
		Idle:        1,
		IdleTimeout: xtime.Duration(90 * time.Second),
		//WaitTimeout: xtime.Duration(10 * time.Millisecond),
		Wait: false,
	}
	pool := NewSlice(config)
	pool.New = func(ctx context.Context) (io.Closer, error) {
		return &closer{}, nil
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 100*time.Millisecond)
	defer cancel()
	conn, err := pool.Get(context.TODO())
	t.Logf("1 pool:%v", pool)
	assert.Nil(t, err)
	_, err = pool.Get(ctx)
	assert.NotNil(t, err)
	t.Logf("err:%s", err)
	pool.Put(context.TODO(), conn, false)
	t.Logf("2 pool:%s", pool)
	_, err = pool.Get(ctx)
	assert.Nil(t, err)
}

func TestSliceStaleClean(t *testing.T) {
	var id = 0
	type connID struct {
		io.Closer
		id int
	}
	config := &Config{
		Active:      1,
		Idle:        1,
		IdleTimeout: xtime.Duration(100 * time.Millisecond),
		//		WaitTimeout: xtime.Duration(10 * time.Millisecond),
		Wait: false,
	}
	pool := NewSlice(config)
	pool.New = func(ctx context.Context) (i io.Closer, err error) {
		id = id + 1
		return &connID{id: id, Closer: &closer{}}, nil
	}
	conn, err := pool.Get(context.TODO())
	defer conn.Close()
	assert.Nil(t, err)
	pool.Put(context.TODO(), conn, false)
	t.Logf("1 pool:%s", pool)
	time.Sleep(200 * time.Millisecond)
	t.Logf("2 pool:%s", pool)
}

func BenchmarkSlice1(b *testing.B) {
	config := &Config{
		Active:      30,
		Idle:        30,
		IdleTimeout: xtime.Duration(90 * time.Second),
		WaitTimeout: xtime.Duration(10 * time.Millisecond),
		Wait:        false,
	}
	pool := NewSlice(config)
	pool.New = func(ctx context.Context) (io.Closer, error) {
		return &closer{}, nil
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn, err := pool.Get(context.TODO())
			if err != nil {
				b.Error(err)
				continue
			}
			c1 := connection{pool: pool, c: conn}
			c1.HandleQuick()
			c1.Close()
		}
	})
}

func BenchmarkSlice2(b *testing.B) {
	config := &Config{
		Active:      30,
		Idle:        30,
		IdleTimeout: xtime.Duration(90 * time.Second),
		WaitTimeout: xtime.Duration(10 * time.Millisecond),
		Wait:        false,
	}
	pool := NewSlice(config)
	pool.New = func(ctx context.Context) (io.Closer, error) {
		return &closer{}, nil
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn, err := pool.Get(context.TODO())
			if err != nil {
				b.Error(err)
				continue
			}
			c1 := connection{pool: pool, c: conn}
			c1.HandleNormal()
			c1.Close()
		}
	})
}
