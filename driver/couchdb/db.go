package couchdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/couchdb/chttp"
)

type db struct {
	*client
	dbName      string
	forceCommit bool
}

func (d *db) path(path string, query url.Values) string {
	url, _ := url.Parse(d.dbName + "/" + strings.TrimPrefix(path, "/"))
	if query != nil {
		url.RawQuery = query.Encode()
	}
	return url.String()
}

func optionsToParams(opts ...map[string]interface{}) (url.Values, error) {
	params := url.Values{}
	for _, optsSet := range opts {
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
				return nil, fmt.Errorf("Cannot convert type %T to []string", i)
			}
			for _, value := range values {
				params.Add(key, value)
			}
		}
	}
	return params, nil
}

// rowsQuery performs a query that returns a rows iterator.
func (d *db) rowsQuery(ctx context.Context, path string, opts map[string]interface{}) (driver.Rows, error) {
	options, err := optionsToParams(opts)
	if err != nil {
		return nil, err
	}
	resp, err := d.Client.DoReq(ctx, kivik.MethodGet, d.path(path, options), nil)
	if err != nil {
		return nil, err
	}
	if err = chttp.ResponseError(resp); err != nil {
		return nil, err
	}
	return newRows(resp.Body), nil
}

// AllDocsContext returns all of the documents in the database.
func (d *db) AllDocsContext(ctx context.Context, opts map[string]interface{}) (driver.Rows, error) {
	return d.rowsQuery(ctx, "_all_docs", opts)
}

// QueryContext queries a view.
func (d *db) QueryContext(ctx context.Context, ddoc, view string, opts map[string]interface{}) (driver.Rows, error) {
	return d.rowsQuery(ctx, fmt.Sprintf("_design/%s/_view/%s", ddoc, view), opts)
}

// GetContext fetches the requested document.
func (d *db) GetContext(ctx context.Context, docID string, doc interface{}, opts map[string]interface{}) error {
	params, err := optionsToParams(opts)
	if err != nil {
		return err
	}
	_, err = d.Client.DoJSON(ctx, http.MethodGet, d.path(docID, params), &chttp.Options{Accept: "application/json; multipart/mixed"}, doc)
	return err
}

func (d *db) CreateDocContext(ctx context.Context, doc interface{}) (docID, rev string, err error) {
	result := struct {
		ID  string `json:"id"`
		Rev string `json:"rev"`
	}{}
	var jsonErr error
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()
	opts := &chttp.Options{
		Body:        chttp.EncodeBody(doc, &jsonErr, cancel),
		ForceCommit: d.forceCommit,
	}
	_, err = d.Client.DoJSON(ctx, kivik.MethodPost, d.dbName, opts, &result)
	if jsonErr != nil {
		return "", "", jsonErr
	}
	return result.ID, result.Rev, err
}

const (
	prefixDesign = "_design/"
	prefixLocal  = "_local/"
)

// encodeDocID ensures that any '/' chars in the docID are escaped, except
// those found in the strings '_design/' and '_local'
func encodeDocID(docID string) string {
	for _, prefix := range []string{prefixDesign, prefixLocal} {
		if strings.HasPrefix(docID, prefix) {
			return prefix + replaceSlash(strings.TrimPrefix(docID, prefix))
		}
	}
	return replaceSlash(docID)
}

// replaceSlash replaces any '/' char with '%2F'
func replaceSlash(docID string) string {
	return strings.Replace(docID, "/", "%2F", -1)
}

func (d *db) PutContext(ctx context.Context, docID string, doc interface{}) (rev string, err error) {
	var jsonErr error
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()
	opts := &chttp.Options{
		Body:        chttp.EncodeBody(doc, &jsonErr, cancel),
		ForceCommit: d.forceCommit,
	}
	var result struct {
		ID  string `json:"id"`
		Rev string `json:"rev"`
	}
	_, err = d.Client.DoJSON(ctx, kivik.MethodPut, d.path(encodeDocID(docID), nil), opts, &result)
	if err != nil {
		return "", err
	}
	if result.ID != docID {
		// This should never happen; this is mostly for debugging and internal use
		return result.Rev, fmt.Errorf("created document ID (%s) does not match that requested (%s)", result.ID, docID)
	}
	return result.Rev, nil
}

func (d *db) DeleteContext(ctx context.Context, docID, rev string) (string, error) {
	query := url.Values{}
	query.Add("rev", rev)
	opts := &chttp.Options{
		ForceCommit: d.forceCommit,
	}
	resp, err := d.Client.DoReq(ctx, kivik.MethodDelete, d.path(docID, query), opts)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return chttp.GetRev(resp)
}

func (d *db) FlushContext(ctx context.Context) (time.Time, error) {
	result := struct {
		T int64 `json:"instance_start_time,string"`
	}{}
	_, err := d.Client.DoJSON(ctx, kivik.MethodPost, d.path("/_ensure_full_commit", nil), nil, &result)
	return time.Unix(0, 0).Add(time.Duration(result.T) * time.Microsecond), err
}

func (d *db) InfoContext(ctx context.Context) (*driver.DBInfo, error) {
	result := struct {
		driver.DBInfo
		Sizes struct {
			File     int64 `json:"file"`
			External int64 `json:"external"`
			Active   int64 `json:"active"`
		} `json:"sizes"`
		UpdateSeq json.RawMessage `json:"update_seq"`
	}{}
	_, err := d.Client.DoJSON(ctx, kivik.MethodGet, d.dbName, nil, &result)
	info := result.DBInfo
	if result.Sizes.File > 0 {
		info.DiskSize = result.Sizes.File
	}
	if result.Sizes.External > 0 {
		info.ExternalSize = result.Sizes.External
	}
	if result.Sizes.Active > 0 {
		info.ActiveSize = result.Sizes.Active
	}
	info.UpdateSeq = string(bytes.Trim(result.UpdateSeq, `"`))
	return &info, err
}

func (d *db) CompactContext(ctx context.Context) error {
	res, err := d.Client.DoReq(ctx, kivik.MethodPost, d.path("/_compact", nil), nil)
	if err != nil {
		return err
	}
	return chttp.ResponseError(res)
}

func (d *db) CompactViewContext(ctx context.Context, ddocID string) error {
	res, err := d.Client.DoReq(ctx, kivik.MethodPost, d.path("/_compact/"+ddocID, nil), nil)
	if err != nil {
		return err
	}
	return chttp.ResponseError(res)
}

func (d *db) ViewCleanupContext(ctx context.Context) error {
	res, err := d.Client.DoReq(ctx, kivik.MethodPost, d.path("/_view_cleanup", nil), nil)
	if err != nil {
		return err
	}
	return chttp.ResponseError(res)
}

func (d *db) SecurityContext(ctx context.Context) (*driver.Security, error) {
	var sec *driver.Security
	_, err := d.Client.DoJSON(ctx, kivik.MethodGet, d.path("/_security", nil), nil, &sec)
	return sec, err
}

func (d *db) SetSecurityContext(ctx context.Context, security *driver.Security) error {
	var jsonErr error
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()
	opts := &chttp.Options{
		Body: chttp.EncodeBody(security, &jsonErr, cancel),
	}
	res, err := d.Client.DoReq(ctx, kivik.MethodPut, d.path("/_security", nil), opts)
	if jsonErr != nil {
		return jsonErr
	}
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return chttp.ResponseError(res)
}

// RevContext returns the most current rev of the requested document.
func (d *db) RevContext(ctx context.Context, docID string) (rev string, err error) {
	res, err := d.Client.DoError(ctx, http.MethodHead, d.path(docID, nil), nil)
	if err != nil {
		return "", err
	}
	return strings.Trim(res.Header.Get("Etag"), `""`), nil
}

// RevsLimitContext returns the maximum number of document revisions that will
// be tracked.
func (d *db) RevsLimitContext(ctx context.Context) (limit int, err error) {
	_, err = d.Client.DoJSON(ctx, http.MethodGet, d.path("/_revs_limit", nil), nil, &limit)
	return limit, err
}

func (d *db) SetRevsLimitContext(ctx context.Context, limit int) error {
	body, err := json.Marshal(limit)
	if err != nil {
		return err
	}
	_, err = d.Client.DoError(ctx, http.MethodPut, d.path("/_revs_limit", nil), &chttp.Options{Body: bytes.NewBuffer(body)})
	return err
}

func (d *db) CopyContext(ctx context.Context, targetID, sourceID string, options map[string]interface{}) (targetRev string, err error) {
	params, err := optionsToParams(options)
	if err != nil {
		return "", err
	}
	opts := &chttp.Options{
		ForceCommit: d.forceCommit,
		Destination: targetID,
	}
	resp, err := d.Client.DoReq(ctx, kivik.MethodCopy, d.path(sourceID, params), opts)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return chttp.GetRev(resp)
}
