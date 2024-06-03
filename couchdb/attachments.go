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

	"github.com/go-kivik/kivik/v4/couchdb/chttp"
	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

func (d *db) PutAttachment(ctx context.Context, docID string, att *driver.Attachment, options driver.Options) (newRev string, err error) {
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

	chttpOpts := chttp.NewOptions(options)

	opts := map[string]interface{}{}
	options.Apply(opts)
	query, err := optionsToParams(opts)
	if err != nil {
		return "", err
	}
	var response struct {
		Rev string `json:"rev"`
	}
	chttpOpts.Body = att.Content
	chttpOpts.ContentType = att.ContentType
	chttpOpts.Query = query
	err = d.Client.DoJSON(ctx, http.MethodPut, d.path(chttp.EncodeDocID(docID)+"/"+att.Filename), chttpOpts, &response)
	if err != nil {
		return "", err
	}
	return response.Rev, nil
}

func (d *db) GetAttachmentMeta(ctx context.Context, docID, filename string, options driver.Options) (*driver.Attachment, error) {
	resp, err := d.fetchAttachment(ctx, http.MethodHead, docID, filename, options)
	if err != nil {
		return nil, err
	}
	att, err := decodeAttachment(resp)
	return att, err
}

func (d *db) GetAttachment(ctx context.Context, docID, filename string, options driver.Options) (*driver.Attachment, error) {
	resp, err := d.fetchAttachment(ctx, http.MethodGet, docID, filename, options)
	if err != nil {
		return nil, err
	}
	return decodeAttachment(resp)
}

func (d *db) fetchAttachment(ctx context.Context, method, docID, filename string, options driver.Options) (*http.Response, error) {
	if method == "" {
		return nil, errors.New("method required")
	}
	if docID == "" {
		return nil, missingArg("docID")
	}
	if filename == "" {
		return nil, missingArg("filename")
	}
	chttpOpts := chttp.NewOptions(options)

	opts := map[string]interface{}{}
	options.Apply(opts)
	var err error
	chttpOpts.Query, err = optionsToParams(opts)
	if err != nil {
		return nil, err
	}
	resp, err := d.Client.DoReq(ctx, method, d.path(chttp.EncodeDocID(docID)+"/"+filename), chttpOpts)
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
		return "", &internal.Error{Status: http.StatusBadGateway, Err: errors.New("no Content-Type in response")}
	}
	return ctype, nil
}

func getDigest(resp *http.Response) (string, error) {
	etag, ok := chttp.ETag(resp)
	if !ok {
		return "", &internal.Error{Status: http.StatusBadGateway, Err: errors.New("ETag header not found")}
	}
	return etag, nil
}

func (d *db) DeleteAttachment(ctx context.Context, docID, filename string, options driver.Options) (newRev string, err error) {
	if docID == "" {
		return "", missingArg("docID")
	}
	opts := map[string]interface{}{}
	options.Apply(opts)
	if rev, _ := opts["rev"].(string); rev == "" {
		return "", missingArg("rev")
	}
	if filename == "" {
		return "", missingArg("filename")
	}

	chttpOpts := chttp.NewOptions(options)

	chttpOpts.Query, err = optionsToParams(opts)
	if err != nil {
		return "", err
	}
	var response struct {
		Rev string `json:"rev"`
	}

	err = d.Client.DoJSON(ctx, http.MethodDelete, d.path(chttp.EncodeDocID(docID)+"/"+filename), chttpOpts, &response)
	if err != nil {
		return "", err
	}
	return response.Rev, nil
}
