package pouchdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/pouchdb/bindings"
	"github.com/flimzy/kivik/errors"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"
)

type db struct {
	db *bindings.DB

	client *client

	// compacting is set true when compaction begins, and unset when the
	// callback returns.
	compacting bool
}

// SetOption sets a connection-time option by replacing the underling DB
// instance.
func (d *db) SetOption(key string, value interface{}) error {
	// Get the existing options
	opts := d.db.Object.Get("__opts")
	// Then set the new value
	opts.Set(key, value)
	// Then re-establish the connection
	d.db = &bindings.DB{Object: d.client.pouch.Object.New("", opts)}
	return nil
}

func (d *db) AllDocsContext(ctx context.Context, options map[string]interface{}) (driver.Rows, error) {
	result, err := d.db.AllDocs(ctx, options)
	if err != nil {
		return nil, err
	}
	return &rows{
		Object: result,
	}, nil
}

func (d *db) GetContext(ctx context.Context, docID string, doc interface{}, options map[string]interface{}) error {
	body, err := d.db.Get(ctx, docID, options)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &doc)
}

func (d *db) CreateDocContext(ctx context.Context, doc interface{}) (docID, rev string, err error) {
	jsonDoc, err := json.Marshal(doc)
	if err != nil {
		return "", "", err
	}
	jsDoc := js.Global.Get("JSON").Call("parse", string(jsonDoc))
	return d.db.Post(ctx, jsDoc)
}

func (d *db) PutContext(ctx context.Context, docID string, doc interface{}) (rev string, err error) {
	jsonDoc, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	jsDoc := js.Global.Get("JSON").Call("parse", string(jsonDoc))
	if id := jsDoc.Get("_id"); id != js.Undefined {
		if id.String() != docID {
			return "", errors.Status(kivik.StatusBadRequest, "id argument must match _id field in document")
		}
	}
	jsDoc.Set("_id", docID)
	return d.db.Put(ctx, jsDoc)
}

func (d *db) DeleteContext(ctx context.Context, docID, rev string) (newRev string, err error) {
	return d.db.Delete(ctx, map[string]string{
		"_id":  docID,
		"_rev": rev,
	})
}

func (d *db) InfoContext(ctx context.Context) (*driver.DBInfo, error) {
	i, err := d.db.Info(ctx)
	return &driver.DBInfo{
		Name:           i.Name,
		CompactRunning: d.compacting,
		DocCount:       i.DocCount,
		UpdateSeq:      i.UpdateSeq,
	}, err
}

func (d *db) CompactContext(_ context.Context) error {
	d.compacting = true
	go func() {
		defer func() { d.compacting = false }()
		if err := d.db.Compact(); err != nil {
			fmt.Fprintf(os.Stderr, "compaction failed: %s", err)
		}
	}()
	return nil
}

// CompactViewContext is unimplemented for PouchDB
func (d *db) CompactViewContext(_ context.Context, _ string) error {
	return nil
}

func (d *db) ViewCleanupContext(_ context.Context) error {
	d.compacting = true
	go func() {
		defer func() { d.compacting = false }()
		if err := d.db.ViewCleanup(); err != nil {
			fmt.Fprintf(os.Stderr, "view cleanup failed: %s", err)
		}
	}()
	return nil
}

func (d *db) SecurityContext(ctx context.Context) (*driver.Security, error) {
	return nil, kivik.ErrNotImplemented
}

func (d *db) SetSecurityContext(_ context.Context, _ *driver.Security) error {
	return kivik.ErrNotImplemented
}

func (d *db) RevsLimitContext(_ context.Context) (limit int, err error) {
	return d.db.RevsLimit()
}

func (d *db) SetRevsLimitContext(_ context.Context, limit int) error {
	return d.SetOption("revs_limit", limit)
}

func (d *db) PutAttachmentContext(ctx context.Context, docID, rev, filename, contentType string, body io.Reader) (newRev string, err error) {
	result, err := d.db.PutAttachment(ctx, docID, filename, rev, body, contentType)
	if err != nil {
		return "", err
	}
	return result.Get("rev").String(), nil
}

func (d *db) GetAttachmentContext(ctx context.Context, docID, rev, filename string) (cType string, md5sum driver.Checksum, body io.ReadCloser, err error) {
	result, err := d.fetchAttachment(ctx, docID, rev, filename)
	if err != nil {
		return "", driver.Checksum{}, nil, err
	}
	cType, body, err = parseAttachment(result)
	return
}

func (d *db) fetchAttachment(ctx context.Context, docID, rev, filename string) (*js.Object, error) {
	var opts map[string]interface{}
	if rev != "" {
		opts["rev"] = rev
	}
	return d.db.GetAttachment(ctx, docID, filename, opts)
}

func parseAttachment(att *js.Object) (cType string, content io.ReadCloser, err error) {
	defer bindings.RecoverError(&err)
	if jsbuiltin.TypeOf(att.Get("write")) == "function" {
		// This looks like a Buffer object; we're in Node.js
		body := att.Call("toString", "binary").String()
		// It might make sense to wrap the Buffer itself in an io.Reader interface,
		// but since this is only for testing, I'm taking the lazy way out, even
		// though it means slurping an extra copy into memory.
		return "", ioutil.NopCloser(strings.NewReader(body)), nil
	}
	// We're in the browser
	return att.Get("type").String(), &blobReader{Object: att}, nil
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
