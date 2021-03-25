package render

import (
	"github.com/pkg/errors"
	"net/http"
)

type Data struct {
	ContentType string
	Data        [][]byte
}

func (r Data) Render(w http.ResponseWriter) (err error) {
	r.WriteContentType(w)
	for _, d := range r.Data {
		if _, err = w.Write(d); err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	return
}

func (r Data) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, []string{r.ContentType})
}
