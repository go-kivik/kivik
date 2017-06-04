package memory

import (
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
	data        map[string]interface{}
	ID          string
	RevID       int64
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
