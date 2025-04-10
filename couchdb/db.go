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

package couchdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/couchdb/chttp"
	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

type db struct {
	*client
	dbName string
}

var (
	_ driver.DB                   = &db{}
	_ driver.Finder               = &db{}
	_ driver.RevGetter            = &db{}
	_ driver.AttachmentMetaGetter = &db{}
	_ driver.PartitionedDB        = &db{}
	_ driver.Updater              = &db{}
)

func (d *db) path(path string) string {
	url, err := url.Parse(d.dbName + "/" + strings.TrimPrefix(path, "/"))
	if err != nil {
		panic("THIS IS A BUG: d.path failed: " + err.Error())
	}
	return url.String()
}

func optionsToParams(opts ...map[string]interface{}) (url.Values, error) {
	params := url.Values{}
	for _, optsSet := range opts {
		if err := encodeKeys(optsSet); err != nil {
			return nil, err
		}
		for key, i := range optsSet {
			var values []string
			switch v := i.(type) {
			case string:
				values = []string{v}
			case []string:
				values = v
			case bool:
				values = []string{fmt.Sprintf("%t", v)}
			case int, uint, uint8, uint16, uint32, uint64, int8, int16, int32, int64:
				values = []string{fmt.Sprintf("%d", v)}
			default:
				return nil, &internal.Error{Status: http.StatusBadRequest, Err: fmt.Errorf("kivik: invalid type %T for options", i)}
			}
			for _, value := range values {
				params.Add(key, value)
			}
		}
	}
	return params, nil
}

// rowsQuery performs a query that returns a rows iterator.
func (d *db) rowsQuery(ctx context.Context, path string, options driver.Options) (driver.Rows, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)
	payload := make(map[string]interface{})
	if keys := opts["keys"]; keys != nil {
		delete(opts, "keys")
		payload["keys"] = keys
	}
	rowsInit := newRows
	if queries := opts["queries"]; queries != nil {
		rowsInit = newMultiQueriesRows
		delete(opts, "queries")
		payload["queries"] = queries
		// Funny that this works even in CouchDB 1.x. It seems 1.x just ignores
		// extra path elements beyond the view name. So yay for accidental
		// backward compatibility!
		path = filepath.Join(path, "queries")
	}
	query, err := optionsToParams(opts)
	if err != nil {
		return nil, err
	}
	chttpOpts := &chttp.Options{Query: query}
	method := http.MethodGet
	if len(payload) > 0 {
		method = http.MethodPost
		chttpOpts.GetBody = chttp.BodyEncoder(payload)
		chttpOpts.Header = http.Header{
			chttp.HeaderIdempotencyKey: []string{},
		}
	}
	resp, err := d.Client.DoReq(ctx, method, d.path(path), chttpOpts)
	if err != nil {
		return nil, err
	}
	if err = chttp.ResponseError(resp); err != nil {
		return nil, err
	}
	return rowsInit(ctx, resp.Body), nil
}

// AllDocs returns all of the documents in the database.
func (d *db) AllDocs(ctx context.Context, options driver.Options) (driver.Rows, error) {
	reqPath := partPath("_all_docs")
	options.Apply(reqPath)
	return d.rowsQuery(ctx, reqPath.String(), options)
}

// DesignDocs returns all of the documents in the database.
func (d *db) DesignDocs(ctx context.Context, options driver.Options) (driver.Rows, error) {
	return d.rowsQuery(ctx, "_design_docs", options)
}

// LocalDocs returns all of the documents in the database.
func (d *db) LocalDocs(ctx context.Context, options driver.Options) (driver.Rows, error) {
	return d.rowsQuery(ctx, "_local_docs", options)
}

// Query queries a view.
func (d *db) Query(ctx context.Context, ddoc, view string, options driver.Options) (driver.Rows, error) {
	reqPath := partPath(fmt.Sprintf("_design/%s/_view/%s", chttp.EncodeDocID(ddoc), chttp.EncodeDocID(view)))
	options.Apply(reqPath)
	return d.rowsQuery(ctx, reqPath.String(), options)
}

// document represents a single document returned by Get
type document struct {
	id          string
	rev         string
	body        io.ReadCloser
	attachments driver.Attachments

	// read will be non-zero once the document has been read.
	read int32
}

func (d *document) Next(row *driver.Row) error {
	if atomic.SwapInt32(&d.read, 1) > 0 {
		return io.EOF
	}
	row.ID = d.id
	row.Rev = d.rev
	row.Doc = d.body
	row.Attachments = d.attachments
	return nil
}

func (d *document) Close() error {
	atomic.StoreInt32(&d.read, 1)
	return d.body.Close()
}

func (*document) UpdateSeq() string { return "" }
func (*document) Offset() int64     { return 0 }
func (*document) TotalRows() int64  { return 0 }

// Get fetches the requested document.
func (d *db) Get(ctx context.Context, docID string, options driver.Options) (*driver.Document, error) {
	resp, err := d.get(ctx, http.MethodGet, docID, options)
	if err != nil {
		return nil, err
	}
	ct, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, &internal.Error{Status: http.StatusBadGateway, Err: err}
	}

	switch ct {
	case typeJSON, typeMPRelated:
		etag, _ := chttp.ETag(resp)
		doc, err := processDoc(docID, ct, params["boundary"], etag, resp.Body)
		if err != nil {
			return nil, err
		}
		return &driver.Document{
			Rev:         doc.rev,
			Body:        doc.body,
			Attachments: doc.attachments,
		}, nil
	default:
		return nil, &internal.Error{Status: http.StatusBadGateway, Err: fmt.Errorf("kivik: invalid content type in response: %s", ct)}
	}
}

func openRevs(revs []string) kivik.Option {
	encoded, _ := json.Marshal(revs)
	return kivik.Param("open_revs", string(encoded))
}

// TODO: Flesh this out.
func (d *db) OpenRevs(ctx context.Context, docID string, revs []string, options driver.Options) (driver.Rows, error) {
	opts := multiOptions{
		kivik.Option(options),
		openRevs(revs),
	}
	resp, err := d.get(ctx, http.MethodGet, docID, opts)
	if err != nil {
		return nil, err
	}
	ct, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, &internal.Error{Status: http.StatusBadGateway, Err: err}
	}

	switch ct {
	case typeJSON, typeMPRelated:
		etag, _ := chttp.ETag(resp)
		return processDoc(docID, ct, params["boundary"], etag, resp.Body)
	case typeMPMixed:
		boundary := strings.Trim(params["boundary"], "\"")
		if boundary == "" {
			return nil, &internal.Error{Status: http.StatusBadGateway, Err: errors.New("kivik: boundary missing for multipart/related response")}
		}
		mpReader := multipart.NewReader(resp.Body, boundary)
		return &multiDocs{
			id:       docID,
			respBody: resp.Body,
			reader:   mpReader,
		}, nil
	default:
		return nil, &internal.Error{Status: http.StatusBadGateway, Err: fmt.Errorf("kivik: invalid content type in response: %s", ct)}
	}
}

func processDoc(docID, ct, boundary, rev string, body io.ReadCloser) (*document, error) {
	switch ct {
	case typeJSON:
		if rev == "" {
			var err error
			body, rev, err = chttp.ExtractRev(body)
			if err != nil {
				return nil, err
			}
		}

		return &document{
			id:   docID,
			rev:  rev,
			body: body,
		}, nil
	case typeMPRelated:
		boundary := strings.Trim(boundary, "\"")
		if boundary == "" {
			return nil, &internal.Error{Status: http.StatusBadGateway, Err: errors.New("kivik: boundary missing for multipart/related response")}
		}
		mpReader := multipart.NewReader(body, boundary)
		body, err := mpReader.NextPart()
		if err != nil {
			return nil, &internal.Error{Status: http.StatusBadGateway, Err: err}
		}

		// TODO: Use a TeeReader here, to avoid slurping the entire body into memory at once
		content, err := io.ReadAll(body)
		if err != nil {
			return nil, &internal.Error{Status: http.StatusBadGateway, Err: err}
		}
		_, rev, err := chttp.ExtractRev(io.NopCloser(bytes.NewReader(content)))
		if err != nil {
			return nil, err
		}

		var metaDoc struct {
			Attachments map[string]attMeta `json:"_attachments"`
		}
		if err := json.Unmarshal(content, &metaDoc); err != nil {
			return nil, &internal.Error{Status: http.StatusBadGateway, Err: err}
		}

		return &document{
			id:   docID,
			rev:  rev,
			body: io.NopCloser(bytes.NewBuffer(content)),
			attachments: &multipartAttachments{
				content:  body,
				mpReader: mpReader,
				meta:     metaDoc.Attachments,
			},
		}, nil
	default:
		return nil, &internal.Error{Status: http.StatusBadGateway, Err: fmt.Errorf("kivik: invalid content type in response: %s", ct)}
	}
}

type multiDocs struct {
	id           string
	respBody     io.Closer
	reader       *multipart.Reader
	readerCloser io.Closer
}

var _ driver.Rows = (*multiDocs)(nil)

func (d *multiDocs) Next(row *driver.Row) error {
	if d.readerCloser != nil {
		if err := d.readerCloser.Close(); err != nil {
			return err
		}
		d.readerCloser = nil
	}
	part, err := d.reader.NextPart()
	if err != nil {
		return err
	}
	ct, params, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	if _, ok := params["error"]; ok {
		var body struct {
			Rev string `json:"missing"`
		}
		err := json.NewDecoder(part).Decode(&body)
		row.ID = d.id
		row.Error = &internal.Error{Status: http.StatusNotFound, Err: errors.New("missing")}
		row.Rev = body.Rev
		return err
	}

	doc, err := processDoc(d.id, ct, params["boundary"], "", part)
	if err != nil {
		return err
	}

	row.ID = doc.id
	row.Doc = doc.body
	row.Rev = doc.rev
	row.Attachments = doc.attachments

	return nil
}

func (d *multiDocs) Close() error {
	if d.readerCloser != nil {
		if err := d.readerCloser.Close(); err != nil {
			return err
		}
		d.readerCloser = nil
	}
	return d.respBody.Close()
}

func (*multiDocs) UpdateSeq() string { return "" }
func (*multiDocs) Offset() int64     { return 0 }
func (*multiDocs) TotalRows() int64  { return 0 }

type attMeta struct {
	ContentType string `json:"content_type"`
	Size        *int64 `json:"length"`
	Follows     bool   `json:"follows"`
}

type multipartAttachments struct {
	content  io.ReadCloser
	mpReader *multipart.Reader
	meta     map[string]attMeta
}

var _ driver.Attachments = &multipartAttachments{}

func (a *multipartAttachments) Next(att *driver.Attachment) error {
	part, err := a.mpReader.NextPart()
	switch err {
	case io.EOF:
		return err
	case nil:
		// fall through
	default:
		return &internal.Error{Status: http.StatusBadGateway, Err: err}
	}

	disp, dispositionParams, err := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
	if err != nil {
		return &internal.Error{Status: http.StatusBadGateway, Err: fmt.Errorf("Content-Disposition: %s", err)}
	}
	if disp != "attachment" {
		return &internal.Error{Status: http.StatusBadGateway, Err: fmt.Errorf("Unexpected Content-Disposition: %s", disp)}
	}
	filename := dispositionParams["filename"]

	meta := a.meta[filename]
	if !meta.Follows {
		return &internal.Error{Status: http.StatusBadGateway, Err: fmt.Errorf("File '%s' not in manifest", filename)}
	}

	size := int64(-1)
	if meta.Size != nil {
		size = *meta.Size
	} else if cl, e := strconv.ParseInt(part.Header.Get("Content-Length"), 10, 64); e == nil { // nolint:gomnd
		size = cl
	}

	var cType string
	if ctHeader, ok := part.Header["Content-Type"]; ok {
		cType, _, err = mime.ParseMediaType(ctHeader[0])
		if err != nil {
			return &internal.Error{Status: http.StatusBadGateway, Err: err}
		}
	} else {
		cType = meta.ContentType
	}

	*att = driver.Attachment{
		Filename:        filename,
		Size:            size,
		ContentType:     cType,
		Content:         part,
		ContentEncoding: part.Header.Get("Content-Encoding"),
	}
	return nil
}

func (a *multipartAttachments) Close() error {
	return a.content.Close()
}

// Rev returns the most current rev of the requested document.
func (d *db) GetRev(ctx context.Context, docID string, options driver.Options) (string, error) {
	resp, err := d.get(ctx, http.MethodHead, docID, options)
	if err != nil {
		return "", err
	}
	_ = resp.Body.Close()
	rev, err := chttp.GetRev(resp)
	if err != nil {
		return "", err
	}
	return rev, err
}

type getOptions struct {
	noMultipartGet bool
}

func (d *db) get(ctx context.Context, method, docID string, options driver.Options) (*http.Response, error) {
	if docID == "" {
		return nil, missingArg("docID")
	}

	var getOpts getOptions
	options.Apply(&getOpts)

	opts := map[string]interface{}{}
	options.Apply(opts)

	chttpOpts := chttp.NewOptions(options)

	chttpOpts.Accept = strings.Join([]string{typeMPMixed, typeMPRelated, typeJSON}, ", ")
	var err error
	chttpOpts.Query, err = optionsToParams(opts)
	if err != nil {
		return nil, err
	}
	if getOpts.noMultipartGet {
		chttpOpts.Accept = typeJSON
	}
	resp, err := d.Client.DoReq(ctx, method, d.path(chttp.EncodeDocID(docID)), chttpOpts)
	if err != nil {
		return nil, err
	}
	err = chttp.ResponseError(resp)
	return resp, err
}

func (d *db) CreateDoc(ctx context.Context, doc interface{}, options driver.Options) (docID, rev string, err error) {
	result := struct {
		ID  string `json:"id"`
		Rev string `json:"rev"`
	}{}

	chttpOpts := chttp.NewOptions(options)

	opts := map[string]interface{}{}
	options.Apply(opts)

	path := d.dbName
	if len(opts) > 0 {
		params, e := optionsToParams(opts)
		if e != nil {
			return "", "", e
		}
		path += "?" + params.Encode()
	}

	chttpOpts.Body = chttp.EncodeBody(doc)

	err = d.Client.DoJSON(ctx, http.MethodPost, path, chttpOpts, &result)
	return result.ID, result.Rev, err
}

type putOptions struct {
	NoMultipartPut bool
}

func putOpts(doc interface{}, options driver.Options) (*chttp.Options, error) {
	chttpOpts := chttp.NewOptions(options)
	opts := map[string]interface{}{}
	options.Apply(opts)
	var err error
	chttpOpts.Query, err = optionsToParams(opts)
	if err != nil {
		return nil, err
	}
	var putOpts putOptions
	options.Apply(&putOpts)
	if putOpts.NoMultipartPut {
		if atts, ok := extractAttachments(doc); ok {
			boundary, size, multipartBody, err := newMultipartAttachments(chttp.EncodeBody(doc), atts)
			if err != nil {
				return nil, err
			}
			chttpOpts.Body = multipartBody
			chttpOpts.ContentLength = size
			chttpOpts.ContentType = fmt.Sprintf(typeMPRelated+"; boundary=%q", boundary)
			return chttpOpts, nil
		}
	}
	chttpOpts.Body = chttp.EncodeBody(doc)
	return chttpOpts, nil
}

func (d *db) Put(ctx context.Context, docID string, doc interface{}, options driver.Options) (rev string, err error) {
	if docID == "" {
		return "", missingArg("docID")
	}
	opts, err := putOpts(doc, options)
	if err != nil {
		return "", err
	}
	var result struct {
		ID  string `json:"id"`
		Rev string `json:"rev"`
	}
	err = d.Client.DoJSON(ctx, http.MethodPut, d.path(chttp.EncodeDocID(docID)), opts, &result)
	if err != nil {
		return "", err
	}
	return result.Rev, nil
}

func (d *db) Update(ctx context.Context, ddoc, funcName, docID string, doc interface{}, options driver.Options) (string, error) {
	opts, err := putOpts(doc, options)
	if err != nil {
		return "", err
	}
	var result struct {
		ID  string `json:"id"`
		Rev string `json:"rev"`
	}
	pathParts := make([]string, 0, 5)
	pathParts = append(pathParts, "_design", chttp.EncodeDocID(ddoc), "_update", chttp.EncodeDocID(funcName))
	method := http.MethodPost
	if docID != "" {
		method = http.MethodPut
		pathParts = append(pathParts, chttp.EncodeDocID(docID))
	}
	err = d.Client.DoJSON(ctx, method, d.path(filepath.Join(pathParts...)), opts, &result)
	if err != nil {
		return "", err
	}
	return result.Rev, nil
}

const attachmentsKey = "_attachments"

func extractAttachments(doc interface{}) (*kivik.Attachments, bool) {
	if doc == nil {
		return nil, false
	}
	v := reflect.ValueOf(doc)
	if v.Type().Kind() == reflect.Ptr {
		return extractAttachments(v.Elem().Interface())
	}
	if stdMap, ok := doc.(map[string]interface{}); ok {
		return interfaceToAttachments(stdMap[attachmentsKey])
	}
	if v.Kind() != reflect.Struct {
		return nil, false
	}
	for i := 0; i < v.NumField(); i++ {
		if v.Type().Field(i).Tag.Get("json") == attachmentsKey {
			return interfaceToAttachments(v.Field(i).Interface())
		}
	}
	return nil, false
}

func interfaceToAttachments(i interface{}) (*kivik.Attachments, bool) {
	switch t := i.(type) {
	case kivik.Attachments:
		atts := make(kivik.Attachments, len(t))
		for k, v := range t {
			atts[k] = v
			delete(t, k)
		}
		return &atts, true
	case *kivik.Attachments:
		atts := new(kivik.Attachments)
		*atts = *t
		*t = nil
		return atts, true
	}
	return nil, false
}

// newMultipartAttachments reads a json stream on in, and produces a
// multipart/related output suitable for a PUT request.
func newMultipartAttachments(in io.ReadCloser, att *kivik.Attachments) (boundary string, size int64, content io.ReadCloser, err error) {
	tmp, err := os.CreateTemp("", "kivik-multipart-*")
	if err != nil {
		return "", 0, nil, err
	}
	body := multipart.NewWriter(tmp)
	w := sync.WaitGroup{}
	w.Add(1)
	go func() {
		err = createMultipart(body, in, att)
		e := in.Close()
		if err == nil {
			err = e
		}
		w.Done()
	}()
	w.Wait()
	if e := tmp.Sync(); err == nil {
		err = e
	}
	if info, e := tmp.Stat(); e == nil {
		size = info.Size()
	} else if err == nil {
		err = e
	}
	if _, e := tmp.Seek(0, 0); e != nil && err == nil {
		err = e
	}
	return body.Boundary(),
		size,
		tmp,
		err
}

func createMultipart(w *multipart.Writer, r io.ReadCloser, atts *kivik.Attachments) error {
	doc, err := w.CreatePart(textproto.MIMEHeader{
		"Content-Type": {typeJSON},
	})
	if err != nil {
		return err
	}
	attJSON := replaceAttachments(r, atts)
	if _, e := io.Copy(doc, attJSON); e != nil {
		return e
	}

	// Sort the filenames to ensure order consistent with json.Marshal's ordering
	// of the stubs in the body
	filenames := make([]string, 0, len(*atts))
	for filename := range *atts {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	for _, filename := range filenames {
		att := (*atts)[filename]
		file, err := w.CreatePart(textproto.MIMEHeader{
			// "Content-Type":        {att.ContentType},
			// "Content-Disposition": {fmt.Sprintf(`attachment; filename=%q`, filename)},
			// "Content-Length":      {strconv.FormatInt(att.Size, 10)},
		})
		if err != nil {
			return err
		}
		if _, err := io.Copy(file, att.Content); err != nil {
			return err
		}
		_ = att.Content.Close()
	}

	return w.Close()
}

type lener interface {
	Len() int
}

type stater interface {
	Stat() (os.FileInfo, error)
}

// attachmentSize determines the size of the `in` stream, possibly by reading
// the entire stream first. If att.Size is already set, this function does
// nothing. It attempts the following methods:
//
//  1. Calls `Len()`, if implemented by `in` (i.e. `*bytes.Buffer`)
//  2. Calls `Stat()`, if implemented by `in` (i.e. `*os.File`) then returns
//     the file's size
//  3. If `in` is an io.Seeker, copy the entire contents to io.Discard to
//     determine size, then reset the reader to the beginning.
//  4. Read the entire stream to determine the size, and replace att.Content
//     to be replayed.
func attachmentSize(att *kivik.Attachment) error {
	if att.Size > 0 {
		return nil
	}
	size, r, err := readerSize(att.Content)
	if err != nil {
		return err
	}
	rc, ok := r.(io.ReadCloser)
	if !ok {
		rc = io.NopCloser(r)
	}

	att.Content = rc
	att.Size = size
	return nil
}

func readerSize(in io.Reader) (int64, io.Reader, error) {
	if ln, ok := in.(lener); ok {
		return int64(ln.Len()), in, nil
	}
	if st, ok := in.(stater); ok {
		info, err := st.Stat()
		if err != nil {
			return 0, nil, err
		}
		return info.Size(), in, nil
	}
	if sk, ok := in.(io.Seeker); ok {
		n, err := io.Copy(io.Discard, in)
		if err != nil {
			return 0, nil, err
		}
		_, err = sk.Seek(0, io.SeekStart)
		return n, in, err
	}
	content, err := io.ReadAll(in)
	if err != nil {
		return 0, nil, err
	}
	buf := bytes.NewBuffer(content)
	return int64(buf.Len()), io.NopCloser(buf), nil
}

// NewAttachment is a convenience function, which sets the size of the attachment
// based on content. This is intended for creating attachments to be uploaded
// using multipart/related capabilities of [github.com/go-kivik/kivik/v4.DB.Put].
// The attachment size will be set to the first of the following found:
//
//  1. `size`, if present. Only the first value is considered.
//  2. content.Len(), if implemented (i.e. [bytes.Buffer])
//  3. content.Stat().Size(), if implemented (i.e. [os.File])
//  4. Read the entire content into memory, to determine the size. This can
//     use a lot of memory for large attachments. Please use a file, or
//     specify the size directly instead.
func NewAttachment(filename, contentType string, content io.Reader, size ...int64) (*kivik.Attachment, error) {
	var filesize int64
	if len(size) > 0 {
		filesize = size[0]
	} else {
		var err error
		filesize, content, err = readerSize(content)
		if err != nil {
			return nil, err
		}
	}
	rc, ok := content.(io.ReadCloser)
	if !ok {
		rc = io.NopCloser(content)
	}
	return &kivik.Attachment{
		Filename:    filename,
		ContentType: contentType,
		Content:     rc,
		Size:        filesize,
	}, nil
}

// replaceAttachments reads a JSON stream on in, looking for the _attachments
// key, then replaces its value with the marshaled version of att.
func replaceAttachments(in io.ReadCloser, atts *kivik.Attachments) io.ReadCloser {
	r, w := io.Pipe()
	go func() {
		stubs, err := attachmentStubs(atts)
		if err != nil {
			_ = w.CloseWithError(err)
			_ = in.Close()
			return
		}
		err = copyWithAttachmentStubs(w, in, stubs)
		e := in.Close()
		if err == nil {
			err = e
		}
		_ = w.CloseWithError(err)
	}()
	return r
}

type stub struct {
	ContentType string `json:"content_type"`
	Size        int64  `json:"length"`
}

func (s *stub) MarshalJSON() ([]byte, error) {
	type attJSON struct {
		stub
		Follows bool `json:"follows"`
	}
	att := attJSON{
		stub:    *s,
		Follows: true,
	}
	return json.Marshal(att)
}

func attachmentStubs(atts *kivik.Attachments) (map[string]*stub, error) {
	if atts == nil {
		return nil, nil
	}
	result := make(map[string]*stub, len(*atts))
	for filename, att := range *atts {
		if err := attachmentSize(att); err != nil {
			return nil, err
		}
		result[filename] = &stub{
			ContentType: att.ContentType,
			Size:        att.Size,
		}
	}
	return result, nil
}

// copyWithAttachmentStubs copies r to w, replacing the _attachment value with the
// marshaled version of atts.
func copyWithAttachmentStubs(w io.Writer, r io.Reader, atts map[string]*stub) error {
	dec := json.NewDecoder(r)
	t, err := dec.Token()
	if err == nil {
		if t != json.Delim('{') {
			return &internal.Error{Status: http.StatusBadRequest, Err: fmt.Errorf("expected '{', found '%v'", t)}
		}
	}
	if err != nil {
		if err != io.EOF {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "%v", t); err != nil {
		return err
	}
	first := true
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return &internal.Error{Status: http.StatusBadRequest, Err: err}
		}
		switch tp := t.(type) {
		case string:
			if !first {
				if _, e := w.Write([]byte(",")); e != nil {
					return e
				}
			}
			first = false
			if _, e := fmt.Fprintf(w, `"%s":`, tp); e != nil {
				return e
			}
			var val json.RawMessage
			if e := dec.Decode(&val); e != nil {
				return e
			}
			if tp == attachmentsKey {
				if e := json.NewEncoder(w).Encode(atts); e != nil {
					return e
				}
				// Once we're here, we can just stream the rest of the input
				// unaltered.
				if _, e := io.Copy(w, dec.Buffered()); e != nil {
					return e
				}
				_, e := io.Copy(w, r)
				return e
			}
			if _, e := w.Write(val); e != nil {
				return e
			}
		case json.Delim:
			if tp != json.Delim('}') {
				return fmt.Errorf("expected '}', found '%v'", t)
			}
			if _, err := fmt.Fprintf(w, "%v", t); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *db) Delete(ctx context.Context, docID string, options driver.Options) (string, error) {
	if docID == "" {
		return "", missingArg("docID")
	}
	opts := map[string]interface{}{}
	options.Apply(opts)
	if rev, _ := opts["rev"].(string); rev == "" {
		return "", missingArg("rev")
	}

	chttpOpts := chttp.NewOptions(options)

	var err error
	chttpOpts.Query, err = optionsToParams(opts)
	if err != nil {
		return "", err
	}

	resp, err := d.Client.DoReq(ctx, http.MethodDelete, d.path(chttp.EncodeDocID(docID)), chttpOpts)
	if err != nil {
		return "", err
	}
	defer chttp.CloseBody(resp.Body)
	if err := chttp.ResponseError(resp); err != nil {
		return "", err
	}
	return chttp.GetRev(resp)
}

func (d *db) Flush(ctx context.Context) error {
	opts := &chttp.Options{
		Header: http.Header{
			chttp.HeaderIdempotencyKey: []string{},
		},
	}
	_, err := d.Client.DoError(ctx, http.MethodPost, d.path("/_ensure_full_commit"), opts)
	return err
}

func (d *db) Compact(ctx context.Context) error {
	opts := &chttp.Options{
		Header: http.Header{
			chttp.HeaderIdempotencyKey: []string{},
		},
	}
	res, err := d.Client.DoReq(ctx, http.MethodPost, d.path("/_compact"), opts)
	if err != nil {
		return err
	}
	defer chttp.CloseBody(res.Body)
	return chttp.ResponseError(res)
}

func (d *db) CompactView(ctx context.Context, ddocID string) error {
	if ddocID == "" {
		return missingArg("ddocID")
	}
	opts := &chttp.Options{
		Header: http.Header{
			chttp.HeaderIdempotencyKey: []string{},
		},
	}
	res, err := d.Client.DoReq(ctx, http.MethodPost, d.path("/_compact/"+ddocID), opts)
	if err != nil {
		return err
	}
	defer chttp.CloseBody(res.Body)
	return chttp.ResponseError(res)
}

func (d *db) ViewCleanup(ctx context.Context) error {
	opts := &chttp.Options{
		Header: http.Header{
			chttp.HeaderIdempotencyKey: []string{},
		},
	}
	res, err := d.Client.DoReq(ctx, http.MethodPost, d.path("/_view_cleanup"), opts)
	if err != nil {
		return err
	}
	defer chttp.CloseBody(res.Body)
	return chttp.ResponseError(res)
}

func (d *db) Security(ctx context.Context) (*driver.Security, error) {
	var sec *driver.Security
	err := d.Client.DoJSON(ctx, http.MethodGet, d.path("/_security"), nil, &sec)
	return sec, err
}

func (d *db) SetSecurity(ctx context.Context, security *driver.Security) error {
	opts := &chttp.Options{
		GetBody: chttp.BodyEncoder(security),
		Header: http.Header{
			chttp.HeaderIdempotencyKey: []string{},
		},
	}
	res, err := d.Client.DoReq(ctx, http.MethodPut, d.path("/_security"), opts)
	if err != nil {
		return err
	}
	defer chttp.CloseBody(res.Body)
	return chttp.ResponseError(res)
}

func (d *db) Copy(ctx context.Context, targetID, sourceID string, options driver.Options) (targetRev string, err error) {
	if sourceID == "" {
		return "", missingArg("sourceID")
	}
	if targetID == "" {
		return "", missingArg("targetID")
	}
	chttpOpts := chttp.NewOptions(options)

	opts := map[string]interface{}{}
	options.Apply(opts)
	chttpOpts.Query, err = optionsToParams(opts)
	if err != nil {
		return "", err
	}
	chttpOpts.Header = http.Header{
		chttp.HeaderDestination: []string{targetID},
	}

	resp, err := d.Client.DoReq(ctx, "COPY", d.path(chttp.EncodeDocID(sourceID)), chttpOpts)
	if err != nil {
		return "", err
	}
	defer chttp.CloseBody(resp.Body)
	if err := chttp.ResponseError(resp); err != nil {
		return "", err
	}
	return chttp.GetRev(resp)
}

func (d *db) Purge(ctx context.Context, docMap map[string][]string) (*driver.PurgeResult, error) {
	result := &driver.PurgeResult{}
	options := &chttp.Options{
		GetBody: chttp.BodyEncoder(docMap),
		Header: http.Header{
			chttp.HeaderIdempotencyKey: []string{},
		},
	}
	err := d.Client.DoJSON(ctx, http.MethodPost, d.path("_purge"), options, &result)
	return result, err
}

var _ driver.RevsDiffer = &db{}

func (d *db) RevsDiff(ctx context.Context, revMap interface{}) (driver.Rows, error) {
	options := &chttp.Options{
		GetBody: chttp.BodyEncoder(revMap),
		Header: http.Header{
			chttp.HeaderIdempotencyKey: []string{},
		},
	}
	resp, err := d.Client.DoReq(ctx, http.MethodPost, d.path("_revs_diff"), options)
	if err != nil {
		return nil, err
	}
	if err = chttp.ResponseError(resp); err != nil {
		return nil, err
	}
	return newRevsDiffRows(ctx, resp.Body), nil
}

type revsDiffParser struct{}

func (p *revsDiffParser) decodeItem(i interface{}, dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}
	var value json.RawMessage
	if err := dec.Decode(&value); err != nil {
		return err
	}
	row := i.(*driver.Row)
	row.ID = t.(string)
	row.Value = bytes.NewReader(value)
	return nil
}

func newRevsDiffRows(ctx context.Context, in io.ReadCloser) driver.Rows {
	iter := newIter(ctx, nil, "", in, &revsDiffParser{})
	iter.objMode = true
	return &rows{iter: iter}
}

// Close is a no-op for the CouchDB driver.
func (db) Close() error { return nil }
