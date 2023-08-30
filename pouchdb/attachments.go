// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package pouchdb

import (
	"context"
	"io"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/pouchdb/v4/bindings"
)

func (d *db) PutAttachment(ctx context.Context, docID string, att *driver.Attachment, options map[string]interface{}) (newRev string, err error) {
	rev, _ := options["rev"].(string)
	result, err := d.db.PutAttachment(ctx, docID, att.Filename, rev, att.Content, att.ContentType)
	if err != nil {
		return "", err
	}
	return result.Get("rev").String(), nil
}

func (d *db) GetAttachment(ctx context.Context, docID, filename string, opts map[string]interface{}) (*driver.Attachment, error) {
	result, err := d.db.GetAttachment(ctx, docID, filename, opts)
	if err != nil {
		return nil, err
	}
	return parseAttachment(result)
}

func parseAttachment(obj *js.Object) (att *driver.Attachment, err error) {
	defer bindings.RecoverError(&err)
	if jsbuiltin.TypeOf(obj.Get("write")) == jsbuiltin.TypeFunction {
		// This looks like a Buffer object; we're in Node.js
		body := obj.Call("toString", "binary").String()
		// It might make sense to wrap the Buffer itself in an io.Reader interface,
		// but since this is only for testing, I'm taking the lazy way out, even
		// though it means slurping an extra copy into memory.
		return &driver.Attachment{
			Content: ioutil.NopCloser(strings.NewReader(body)),
		}, nil
	}
	// We're in the browser
	return &driver.Attachment{
		ContentType: obj.Get("type").String(),
		Content:     &blobReader{Object: obj},
	}, nil
}

type blobReader struct {
	*js.Object
	offset int
	Size   int `js:"size"`
}

var _ io.ReadCloser = &blobReader{}

func (b *blobReader) Read(p []byte) (n int, err error) {
	defer bindings.RecoverError(&err)
	if b.offset >= b.Size {
		return 0, io.EOF
	}
	end := b.offset + len(p) + 1 // end is the first byte not included, not the last byte included, so add 1
	if end > b.Size {
		end = b.Size
	}
	slice := b.Call("slice", b.offset, end)
	fileReader := js.Global.Get("FileReader").New()
	var wg sync.WaitGroup
	wg.Add(1)
	fileReader.Set("onload", js.MakeFunc(func(this *js.Object, _ []*js.Object) interface{} {
		defer wg.Done()
		n = copy(p, js.Global.Get("Uint8Array").New(this.Get("result")).Interface().([]uint8))
		return nil
	}))
	fileReader.Call("readAsArrayBuffer", slice)
	wg.Wait()
	b.offset += n
	return
}

func (b *blobReader) Close() (err error) {
	defer bindings.RecoverError(&err)
	b.Call("close")
	return nil
}

func (d *db) DeleteAttachment(ctx context.Context, docID, filename string, options map[string]interface{}) (newRev string, err error) {
	rev, _ := options["rev"].(string)
	result, err := d.db.RemoveAttachment(ctx, docID, filename, rev)
	if err != nil {
		return "", err
	}
	return result.Get("rev").String(), nil
}
