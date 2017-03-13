package serve

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

//go:generate go-bindata -pkg serve -nocompress -prefix files -o files.go files

func favicon(w http.ResponseWriter, r *http.Request) error {
	s := GetService(r)
	var ico io.Reader
	if s.Favicon == "" {
		asset, err := Asset("favicon.ico")
		if err != nil {
			panic(err)
		}
		ico = bytes.NewBuffer(asset)
	} else {
		file, err := os.Open(s.Favicon)
		if err != nil {
			if os.IsNotExist(err) {
				return errors.Status(kivik.StatusNotFound, "not found")
			}
			return err
		}
		ico = file
		defer file.Close()
	}
	w.Header().Set("Content-Type", "image/x-icon")
	_, err := io.Copy(w, ico)
	return err
}
