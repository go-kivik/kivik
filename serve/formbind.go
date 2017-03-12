package serve

import (
	"encoding/json"
	"mime"
	"net/http"

	"github.com/ajg/form"
	"github.com/pkg/errors"
)

// BindParams binds the request form or JSON body to the provided struct.
func BindParams(r *http.Request, i interface{}) error {
	mtype, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch mtype {
	case typeJSON:
		defer r.Body.Close()
		return json.NewDecoder(r.Body).Decode(i)
	case typeForm:
		defer r.Body.Close()
		return form.NewDecoder(r.Body).Decode(i)
	}
	return errors.Errorf("unable to bind media type %s", mtype)
}
