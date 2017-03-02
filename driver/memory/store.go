package memory

import "sync"

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
	Rev         string
	Attachments map[string]file
}

type database struct {
	mutex     sync.RWMutex
	docs      map[string]*document
	updateSeq int64
}

func (d *db) getDB() *database {
	c := d.client
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	database, _ := c.dbs[d.dbName]
	return database
}
