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
	"context"
	"errors"
	"net/http"

	"github.com/go-kivik/couchdb/v4/chttp"
	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

func (d *db) PutAttachment(ctx context.Context, docID string, att *driver.Attachment, options map[string]interface{}) (newRev string, err error) {
	if docID == "" {
		return "", missingArg("docID")
	}
	if att == nil {
		return "", missingArg("att")
	}
	if att.Filename == "" {
		return "", missingArg("att.Filename")
	}
	if att.Content == nil {
		return "", missingArg("att.Content")
	}

	opts, err := chttp.NewOptions(options)
	if err != nil {
		return "", err
	}

	query, err := optionsToParams(options)
	if err != nil {
		return "", err
	}
	var response struct {
		Rev string `json:"rev"`
	}
	opts.Body = att.Content
	opts.ContentType = att.ContentType
	opts.Query = query
	err = d.Client.DoJSON(ctx, http.MethodPut, d.path(chttp.EncodeDocID(docID)+"/"+att.Filename), opts, &response)
	if err != nil {
		return "", err
	}
	return response.Rev, nil
}

func (d *db) GetAttachmentMeta(ctx context.Context, docID, filename string, options map[string]interface{}) (*driver.Attachment, error) {
	resp, err := d.fetchAttachment(ctx, http.MethodHead, docID, filename, options)
	if err != nil {
		return nil, err
	}
	att, err := decodeAttachment(resp)
	return att, err
}

func (d *db) GetAttachment(ctx context.Context, docID, filename string, options map[string]interface{}) (*driver.Attachment, error) {
	resp, err := d.fetchAttachment(ctx, http.MethodGet, docID, filename, options)
	if err != nil {
		return nil, err
	}
	return decodeAttachment(resp)
}

func (d *db) fetchAttachment(ctx context.Context, method, docID, filename string, options map[string]interface{}) (*http.Response, error) {
	if method == "" {
		return nil, errors.New("method required")
	}
	if docID == "" {
		return nil, missingArg("docID")
	}
	if filename == "" {
		return nil, missingArg("filename")
	}
	opts, err := chttp.NewOptions(options)
	if err != nil {
		return nil, err
	}

	opts.Query, err = optionsToParams(options)
	if err != nil {
		return nil, err
	}
	resp, err := d.Client.DoReq(ctx, method, d.path(chttp.EncodeDocID(docID)+"/"+filename), opts)
	if err != nil {
		return nil, err
	}
	return resp, chttp.ResponseError(resp)
}

func decodeAttachment(resp *http.Response) (*driver.Attachment, error) {
	cType, err := getContentType(resp)
	if err != nil {
		return nil, err
	}
	digest, err := getDigest(resp)
	if err != nil {
		return nil, err
	}

	return &driver.Attachment{
		ContentType: cType,
		Digest:      digest,
		Size:        resp.ContentLength,
		Content:     resp.Body,
	}, nil
}

func getContentType(resp *http.Response) (string, error) {
	ctype := resp.Header.Get("Content-Type")
	if _, ok := resp.Header["Content-Type"]; !ok {
		return "", &kivik.Error{Status: http.StatusBadGateway, Err: errors.New("no Content-Type in response")}
	}
	return ctype, nil
}

func getDigest(resp *http.Response) (string, error) {
	etag, ok := chttp.ETag(resp)
	if !ok {
		return "", &kivik.Error{Status: http.StatusBadGateway, Err: errors.New("ETag header not found")}
	}
	return etag, nil
}

func (d *db) DeleteAttachment(ctx context.Context, docID, filename string, options map[string]interface{}) (newRev string, err error) {
	if docID == "" {
		return "", missingArg("docID")
	}
	if rev, _ := options["rev"].(string); rev == "" {
		return "", missingArg("rev")
	}
	if filename == "" {
		return "", missingArg("filename")
	}

	opts, err := chttp.NewOptions(options)
	if err != nil {
		return "", err
	}

	opts.Query, err = optionsToParams(options)
	if err != nil {
		return "", err
	}
	var response struct {
		Rev string `json:"rev"`
	}

	err = d.Client.DoJSON(ctx, http.MethodDelete, d.path(chttp.EncodeDocID(docID)+"/"+filename), opts, &response)
	if err != nil {
		return "", err
	}
	return response.Rev, nil
}
