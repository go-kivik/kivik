package memory

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type file struct {
	ContentType string
	Data        []byte
}

type document struct {
	revs []*revision
}

type revision struct {
	data        []byte
	ID          int64
	Rev         string
	Attachments map[string]file
}

type database struct {
	mutex     sync.RWMutex
	docs      map[string]*document
	updateSeq int64
}

var rnd *rand.Rand
var rndMU = &sync.Mutex{}

func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (d *database) latestRevision(docID string) (*revision, bool) {
	doc, ok := d.docs[docID]
	if ok {
		last := doc.revs[len(doc.revs)-1]
		return last, true
	}
	return nil, false
}

func (d *database) addRevision(docID string, data []byte, att map[string]file) string {
	if d.docs[docID] == nil {
		d.docs[docID] = &document{
			revs: make([]*revision, 0, 1),
		}
	}
	var revID int64
	l := len(d.docs[docID].revs)
	if l == 0 {
		revID = 1
	} else {
		revID = d.docs[docID].revs[l-1].ID + 1
	}
	revStr := randStr()
	d.docs[docID].revs = append(d.docs[docID].revs, &revision{
		data:        data,
		ID:          revID,
		Rev:         revStr,
		Attachments: att,
	})
	return fmt.Sprintf("%d-%s", revID, revStr)
}
