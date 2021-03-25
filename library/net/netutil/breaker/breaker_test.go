package breaker

import (
	"errors"
	"github.com/stretchr/testify/assert"
	xtime "github.com/zombie-k/kylin/library/time"
	"testing"
	"time"
)

func TestGroup(t *testing.T) {
	g1 := NewGroup(nil)
	g2 := NewGroup(_conf)
	t.Logf("g1:%+v", g1)
	t.Logf("g2:%+v", g2)
	assert.Equal(t, g1.conf, g2.conf)

	brk := g2.Get("key")
	brk1 := g2.Get("key1")
	brk2 := g2.Get("key")
	assert.Equal(t, brk, brk2)
	t.Logf("brk:%+v", brk)
	t.Logf("brk1:%+v", brk1)
	t.Logf("brk2:%+v", brk2)
	t.Logf("g2:%+v", g2)

	g := NewGroup(_conf)
	c := &Config{
		Window:    xtime.Duration(1 * time.Second),
		Bucket:    10,
		Request:   100,
		SwitchOff: !_conf.SwitchOff,
	}
	g.Reload(c)
	assert.NotEqual(t, g.conf.SwitchOff, _conf.SwitchOff)
	t.Logf("g:%+v", g)
}

func TestInit(t *testing.T) {
	switchOff := _conf.SwitchOff
	c := &Config{
		SwitchOff: !switchOff,
		Window:    xtime.Duration(3 * time.Second),
		Bucket:    20,
		Request:   200,
	}
	Init(c)
	t.Logf("c:%+v", c)
}

func TestGo(t *testing.T) {
	err := Go("test_run", func() error {
		t.Log("breaker allow, callback run()")
		return nil
	}, func() error {
		t.Log("breaker not allow, callback fallback()")
		return errors.New("breaker not allow")
	})
	t.Log(err)

	_group.Reload(&Config{
		SwitchOff: true,
		Window:    xtime.Duration(3 * time.Second),
		Bucket:    10,
		Request:   100,
	})

	err = Go("test_fallback", func() error {
		t.Log("breaker allow, callback run()")
		return nil
	}, func() error {
		t.Log("breaker not allow, callback fallback()")
		return errors.New("breaker not allow")
	})
	t.Log(err)
}
