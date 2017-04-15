package couchdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/couchdb/chttp"
)

type replicationError struct {
	status int
	reason string
}

func (re *replicationError) Error() string {
	return re.reason
}

func (re *replicationError) StatusCode() int {
	return re.status
}

func (re *replicationError) UnmarshalJSON(data []byte) error {
	reason := bytes.Trim(data, `"`)
	re.reason = string(reason)
	parts := bytes.SplitN(reason, []byte(":"), 2)
	switch string(parts[0]) {
	case "db_not_found":
		re.status = kivik.StatusNotFound
	default:
		re.status = kivik.StatusInternalServerError
	}
	return nil
}

type replication struct {
	DocID string    `json:"_id"`
	RepID string    `json:"_replication_id"`
	Src   string    `json:"source"`
	Tgt   string    `json:"target"`
	Start time.Time `json:"-"`
	Ste   string    `json:"_replication_state"`
	Err   error     `json:"-"`
	*db
}

var _ driver.Replication = &replication{}

func (r *replication) ReplicationID() string { return r.RepID }
func (r *replication) Source() string        { return r.Src }
func (r *replication) Target() string        { return r.Tgt }
func (r *replication) StartTime() time.Time  { return r.Start }

func (r *replication) Cancel(ctx context.Context) error {
	var doc map[string]interface{}
	if err := r.Ge(ctx, r.DocID, &doc, nil); err != nil {
		return err
	}
	doc["cancel"] = true
	_, err := r.Put(ctx, r.DocID, doc)
	return err
}

func (r *replication) Update(ctx context.Context, state *driver.ReplicationState) error {
	reps, err := r.db.client.GetReplications(ctx, map[string]interface{}{"key": `"` + r.DocID + `"`})
	if err != nil {
		return err
	}
	fmt.Printf("len(reps) = %d\n", len(reps))
	if len(reps) == 0 {
		return errors.New("replication not found in _replicator database")
	}
	if len(reps) > 1 {
		panic("too many replications returned")
	}
	fmt.Printf("rep id = %s\n", reps[0].ReplicationID())
	if state.ReplicationID == "" {
		state.ReplicationID = reps[0].ReplicationID()
		// state.Source = reps[0].Source()
		// state.Target = reps[0].Target()
	}
	/*
		type ReplicationState struct {
			DocWriteFailures int64     `json:"doc_write_failures"`
			DocsRead         int64     `json:"docs_read"`
			DocsWritten      int64     `json:"docs_written"`
			Progress         float64   `json:"progress"`
			LastUpdate       time.Time // updated_on / time.Now() for pouchdb
			EndTime          time.Time `json:"end_time"`
			SourceSeq        string    `json:"source_seq"` // last_seq for PouchDB
			Status           string    `json:"status"`
			Error            error     `json:"-"`
			UpdateFunc       func(*ReplicationState) error
		}
		type Replication interface {
			Source() string
			Target() string
			// Update should fetch the current replication state from the server.
			Update(context.Context, *ReplicationState) error
			Cancel(context.Context) error
			Delete(context.Context) error
		}
	*/
	spew.Dump(reps)
	return nil
}

func (r *replication) Delete(ctx context.Context) error {
	rev, err := r.Rev(ctx, r.DocID)
	if err != nil {
		return err
	}
	_, err = r.db.Delete(ctx, r.DocID, rev)
	return err
}

func (c *client) GetReplications(ctx context.Context, options map[string]interface{}) ([]driver.Replication, error) {
	delete(options, "conflicts")
	delete(options, "update_seq")
	options["include_docs"] = true
	params, err := optionsToParams(options)
	if err != nil {
		return nil, err
	}
	var result struct {
		Rows []struct {
			Doc struct {
				DocID         string            `json:"_id"`
				ReplicationID string            `json:"_replication_id"`
				Source        string            `json:"source"`
				Target        string            `json:"target"`
				State         string            `json:"_replication_state"`
				Error         *replicationError `json:"_replication_state_reason,omitempty"`
			} `json:"doc"`
		} `json:"rows"`
	}
	path := "/_replicator/_all_docs"
	if params != nil {
		path += "?" + params.Encode()
	}
	if _, err = c.DoJSON(ctx, kivik.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	reps := make([]driver.Replication, 0, len(result.Rows))
	for _, row := range result.Rows {
		if row.Doc.DocID == "_design/_replicator" {
			continue
		}
		rep := &replication{
			DocID: row.Doc.DocID,
			RepID: row.Doc.ReplicationID,
			Src:   row.Doc.Source,
			Tgt:   row.Doc.Target,
			Ste:   row.Doc.State,
			Err:   row.Doc.Error,
		}
		reps = append(reps, rep)
	}
	return reps, nil
}

func (c *client) Replicate(ctx context.Context, targetDSN, sourceDSN string, options map[string]interface{}) (driver.Replication, error) {
	if options == nil {
		options = make(map[string]interface{})
	}
	// Allow overriding source and target with options, i.e. for OAuth1 options
	if _, ok := options["source"]; !ok {
		options["source"] = sourceDSN
	}
	if _, ok := options["target"]; !ok {
		options["target"] = targetDSN
	}
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(options); err != nil {
		return nil, err
	}
	var repStub struct {
		ID string `json:"id"`
	}
	resp, err := c.Client.DoJSON(ctx, kivik.MethodPost, "/_replicator", &chttp.Options{Body: body}, &repStub)
	if err != nil {
		return nil, err
	}
	rep := &replication{
		DocID: repStub.ID,
	}
	rep.Start, _ = time.Parse(time.RFC1123, resp.Header.Get("Date"))
	rep.db = &db{client: c, dbName: "_replicator", forceCommit: true}

	return rep, nil
}
