package test

import (
	"testing"

	_ "github.com/go-kivik/kivik/v4/x/memorydb"
)

func init() {
	RegisterMemoryDBSuite()
}

func TestMemory(t *testing.T) {
	MemoryTest(t)
}
