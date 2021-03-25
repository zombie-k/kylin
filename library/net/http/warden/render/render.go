package render

import "net/http"

type Render interface {
	// Render render it to http response writer.
	Render(w http.ResponseWriter) error
	// WriteContentType write content-type to http response writer.
	WriteContentType(w http.ResponseWriter)
}

var (
	_ Render = &JSON{}
	_ Render = &MapJSON{}
	_ Render = &String{}
	_ Render = &Data{}
)

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) != 0 {
		header["Content-Type"] = value
	}
}
