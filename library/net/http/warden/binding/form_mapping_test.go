package binding

import (
	"fmt"
	"reflect"
	"testing"
)

type card struct {
	Ids string `json:"ids" form:"ids" default:"1"`
}
type person struct {
	Name string `json:"name" form:"name"`
	ID   int    `json:"id" form:"id"`
	Sex  string `json:"sex" form:"sex"`
	//Cards []card `json:"cards" form:"cards" default:"abc,def,igh"`
	Love  []string `json:"love" form:"love" default:"a,b,c,d"`
	Cards []card   `json:"cards" form:"cards,split" default:"1,2,3,4"`
}

func TestTest(t *testing.T) {
	s := new(person)
	t1 := reflect.TypeOf(s)
	fmt.Println(t1)
	tp := t1.Elem()
	fmt.Println(tp)
	for i := 0; i < tp.NumField(); i++ {
		if defV := tp.Field(i).Tag.Get("default"); defV != "" {
			dv := tp.Field(i).Type.Elem()
			fmt.Println(dv, dv.Kind())
		}
	}
}

func TestTest2(t *testing.T) {
	p := new(person)
	obj := reflect.TypeOf(p)
	tp := obj.Elem()
	for i := 0; i < tp.NumField(); i++ {
		tag := tp.Field(i).Tag.Get("form")
		_, option := parseTag(tag)
		if defV := tp.Field(i).Tag.Get("default"); defV != "" {
			fmt.Println(defV)
			dv := reflect.New(tp.Field(i).Type).Elem()
			setWithProperType(tp.Field(i).Type.Kind(), []string{defV}, dv, option)
			fmt.Println(dv)
		}
	}
}

func TestTest3(t *testing.T) {
	var form = map[string][]string{
		"name": []string{"Lily"},
		"id":   []string{"12345"},
		"sex":  []string{"female"},
		"ids":  []string{"23"},
	}
	/*
		p := &person{
			Name:  "Anna",
			ID:    111111,
			Sex:   "female",
			Cards: []string{"a", "b", "c"},
		}
	*/
	p1 := &person{}
	mapForm(p1, form)
	fmt.Println(p1)

}
