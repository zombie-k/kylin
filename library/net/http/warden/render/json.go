package render

import (
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"net/http"
)

var jsonContentType = []string{"application/json; charset=utf-8"}

// JSON common json struct
type JSON struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, value interface{}) (err error) {
	var jsonBytes []byte
	writeContentType(w, jsonContentType)
	if jsonBytes, err = jsoniter.Marshal(value); err != nil {
		err = errors.WithStack(err)
		return
	}
	if _, err = w.Write(jsonBytes); err != nil {
		err = errors.WithStack(err)
	}
	return
}

func (r JSON) Render(w http.ResponseWriter) error {
	return writeJSON(w, r)
}

func (r JSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

type MapJSON map[string]interface{}

func (m MapJSON) Render(w http.ResponseWriter) error {
	return writeJSON(w, m)
}

func (m MapJSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}
