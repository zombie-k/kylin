package binding

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net/http"
	"testing"
)

type FooStruct struct {
	Foo string `msgpack:"foo" json:"foo" form:"foo" xml:"foo" validate:"required"`
	//Foo string `msgpack:"foo" json:"foo" form:"foo" xml:"foo"`
}

type FooBarStruct struct {
	FooStruct
	Bar   string   `msgpack:"bar" json:"bar" form:"bar" xml:"bar" validate:"required"`
	Slice []string `form:"slice" validate:"max=10"`
}

type Int8SliceStruct struct {
	State []int8 `form:"state,split"`
}

type Int64SliceStruct struct {
	State []int64 `form:"state,split"`
}

type StringSliceStruct struct {
	State []string `form:"state,split"`
}

func TestBindingDefault(t *testing.T) {
	assert.Equal(t, Default(http.MethodGet, ""), Form)
	assert.Equal(t, Default(http.MethodGet, MIMEJson), Form)

	assert.Equal(t, Default(http.MethodPost, MIMEJson), JSON)
	assert.Equal(t, Default(http.MethodPut, MIMEJson), JSON)
	assert.Equal(t, Default(http.MethodPost, MIMEJson+"; charset=utf-8"), JSON)
	assert.Equal(t, Default(http.MethodPut, MIMEJson+"; charset=utf-8"), JSON)

	assert.Equal(t, Default(http.MethodPost, MIMEXml), XML)
	assert.Equal(t, Default(http.MethodPut, MIMEXml2), XML)

	assert.Equal(t, Default(http.MethodPost, MIMEPOSTForm), Form)
	assert.Equal(t, Default(http.MethodPut, MIMEPOSTForm+"; charset=utf-8"), Form)
	assert.Equal(t, Default(http.MethodPut, MIMEMultipartPOSTForm+"; charset=utf-8"), Form)
}

func TestStripContentType(t *testing.T) {
	c1 := "application/json+xml"
	c2 := "application/json+xml; charset=utf-8"
	assert.Equal(t, "application/json+xml", stripContentTypeParam(c1))
	assert.Equal(t, "application/json+xml", stripContentTypeParam(c2))
}

func TestBindInt8Form(t *testing.T) {
	params := "state=1,2,3"
	req, _ := http.NewRequest(http.MethodGet, "http://demo.server/test?"+params, nil)
	q := new(Int8SliceStruct)
	Form.Bind(req, q)
	assert.Equal(t, []int8{1, 2, 3}, q.State)

	params = "state=1,2,3,256"
	req, _ = http.NewRequest(http.MethodGet, "http://demo.server/test?"+params, nil)
	q = new(Int8SliceStruct)
	assert.Error(t, Form.Bind(req, q))

	params = "state="
	req, _ = http.NewRequest(http.MethodGet, "http://demo.server/test?"+params, nil)
	q = new(Int8SliceStruct)
	assert.NoError(t, Form.Bind(req, q))
	assert.Len(t, q.State, 0)

	params = "state=1,,2"
	req, _ = http.NewRequest(http.MethodGet, "http://demo.server/test?"+params, nil)
	q = new(Int8SliceStruct)
	assert.NoError(t, Form.Bind(req, q))
	assert.Equal(t, []int8{1, 2}, q.State)
}

func TestBindInt64Form(t *testing.T) {
	params := "state=1,,2,3"
	req, _ := http.NewRequest(http.MethodGet, "http://demo.server/test?"+params, nil)
	q := new(Int64SliceStruct)
	assert.NoError(t, Form.Bind(req, q))
	assert.Equal(t, []int64{1, 2, 3}, q.State)

	params = "state="
	req, _ = http.NewRequest(http.MethodGet, "http://demo.server/test?"+params, nil)
	q = new(Int64SliceStruct)
	assert.NoError(t, Form.Bind(req, q))
	assert.Equal(t, []int64{}, q.State)
	assert.Len(t, q.State, 0)
}

func TestBindStringForm(t *testing.T) {
	params := "state=1,2,,,3"
	req, _ := http.NewRequest(http.MethodGet, "http://demo.server/test?"+params, nil)
	q := new(StringSliceStruct)
	assert.NoError(t, Form.Bind(req, q))
	assert.Equal(t, []string{"1", "2", "3"}, q.State)

	params = "state="
	req, _ = http.NewRequest(http.MethodGet, "http://demo.server/test?"+params, nil)
	q = new(StringSliceStruct)
	assert.NoError(t, Form.Bind(req, q))
	assert.Equal(t, []string{}, q.State)
	assert.Len(t, q.State, 0)

	params = "state=q,,q,"
	req, _ = http.NewRequest(http.MethodGet, "http://demo.server/test?"+params, nil)
	q = new(StringSliceStruct)
	assert.NoError(t, Form.Bind(req, q))
	assert.Equal(t, []string{"q", "q"}, q.State)
	assert.Len(t, q.State, 2)
}

func TestBindingJSON(t *testing.T) {
	assert.Equal(t, JSON.Name(), "json")
	path := "/"
	badPath := "/"
	body := `{"foo":"bar"}`
	badBody := `{"bar":"foo"}`

	obj := &FooStruct{}
	req, _ := http.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	err := JSON.Bind(req, obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)

	obj = &FooStruct{}
	req, _ = http.NewRequest(http.MethodPost, badPath, bytes.NewBufferString(badBody))
	err = JSON.Bind(req, obj)
	assert.Error(t, err)
}

func TestBindingForm(t *testing.T) {
	method := http.MethodPost
	path := "/?foo=unused&slice=a&slice=b"
	body := "foo=bar&bar=foo"
	badPath := "/"
	badBody := "bar2=foo"

	b := Form
	assert.Equal(t, "form", b.Name())
	obj := &FooBarStruct{}
	req, _ := http.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Add("Content-Type", MIMEPOSTForm)
	err := b.Bind(req, obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "foo", obj.Bar)
	assert.Equal(t, []string{"a", "b"}, obj.Slice)

	obj = &FooBarStruct{}
	req, _ = http.NewRequest(method, badPath, bytes.NewBufferString(badBody))
	err = JSON.Bind(req, obj)
	assert.Error(t, err)
}

func TestBindingForm2(t *testing.T) {
	method := http.MethodGet
	path := "/?foo=bar&bar=foo&slice=a&slice=b"
	body := "foo=unused"
	badPath := "/?bar2=foo"
	badBody := ""

	b := Form
	obj := &FooBarStruct{}
	req, _ := http.NewRequest(method, path, bytes.NewBufferString(body))
	err := b.Bind(req, obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "foo", obj.Bar)
	assert.Equal(t, []string{"a", "b"}, obj.Slice)

	obj = &FooBarStruct{}
	req, _ = http.NewRequest(method, badPath, bytes.NewBufferString(badBody))
	err = JSON.Bind(req, obj)
	assert.Error(t, err)
}

func TestBindingQuery(t *testing.T) {
	method := "POST"
	path := "/?foo=bar&bar=foo"
	body := "foo=unused&slice=a&slice=b"

	b := Query
	obj := &FooBarStruct{}
	req, _ := http.NewRequest(method, path, bytes.NewBufferString(body))
	err := b.Bind(req, obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "foo", obj.Bar)
	assert.Len(t, obj.Slice, 0)
}

func TestBindingXml(t *testing.T) {
	b := XML
	assert.Equal(t, "xml", b.Name())

	path := "/"
	body := "<map><foo>bar</foo></map>"
	obj := &FooStruct{}
	req, _ := http.NewRequest("POST", path, bytes.NewBufferString(body))
	err := b.Bind(req, obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)
}

func TestBindingFormPost(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/?foo=getfoo&bar=getbar&slice=abc&slice=123&slice=what", bytes.NewBufferString("foo=bar&bar=foo"))
	req.Header.Set("Content-Type", MIMEPOSTForm)
	var obj FooBarStruct
	FormPost.Bind(req, &obj)
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "foo", obj.Bar)
}

func TestBidingFormMultipart(t *testing.T) {
	boundary := "--testboundary"
	body := new(bytes.Buffer)
	mv := multipart.NewWriter(body)
	mv.SetBoundary(boundary)
	mv.WriteField("foo", "bar")
	mv.WriteField("bar", "foo")
	req, _ := http.NewRequest(http.MethodPost, "/?foo=getfoo&bar=getbar", body)
	req.Header.Set("Content-Type", MIMEMultipartPOSTForm+"; boundary="+boundary)
	mv.Close()

	var obj FooBarStruct
	FormMultipart.Binding(req, &obj)
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "foo", obj.Bar)
}

func TestValidationFails(t *testing.T) {
	var obj FooStruct
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(`{"bar": "foo"}`))
	err := JSON.Bind(req, &obj)
	assert.Error(t, err)
}

func TestValidationDisabled(t *testing.T) {
	backup := Validator
	Validator = nil
	defer func() {
		Validator = backup
	}()

	var obj FooStruct
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(`{"bar": "foo"}`))
	err := JSON.Bind(req, &obj)
	assert.NoError(t, err)
}

func TestExistsSucceeds(t *testing.T) {
	type HogeStruct struct {
		Hoge *int `json:"hoge" binding:"exists"`
	}

	var obj HogeStruct
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(`{"hoge": 0}`))
	err := JSON.Bind(req, &obj)
	assert.NoError(t, err)
}

func TestFormDefaultValue(t *testing.T) {
	type ComplexDefaultStruct struct {
		Int        int     `form:"int" default:"999"`
		String     string  `form:"string" default:"default-string"`
		Bool       bool    `form:"bool" default:"false"`
		Int64Slice []int64 `form:"int64_slice,split" default:"1,2,3,4"`
		Int8Slice  []int8  `form:"int8_slice,split" default:"1,2,3,4"`
	}

	params := "int=333&string=hello&bool=true&int64_slice=5,6,7,8&int8_slice=5,6,7,8"
	req, _ := http.NewRequest("GET", "http://demo.api/test?"+params, nil)
	q := new(ComplexDefaultStruct)
	assert.NoError(t, Form.Bind(req, q))
	assert.Equal(t, 333, q.Int)
	assert.Equal(t, "hello", q.String)
	assert.Equal(t, true, q.Bool)
	assert.Equal(t, []int64{5, 6, 7, 8}, q.Int64Slice)
	assert.EqualValues(t, []int8{5, 6, 7, 8}, q.Int8Slice)

	params = "string=hello&bool=false"
	req, _ = http.NewRequest("GET", "http://demo.api/test?"+params, nil)
	q = new(ComplexDefaultStruct)
	assert.NoError(t, Form.Bind(req, q))
	assert.Equal(t, 999, q.Int)
	assert.Equal(t, "hello", q.String)
	assert.Equal(t, false, q.Bool)
	assert.Equal(t, []int64{1, 2, 3, 4}, q.Int64Slice)
	assert.Equal(t, []int8{1, 2, 3, 4}, q.Int8Slice)
}
