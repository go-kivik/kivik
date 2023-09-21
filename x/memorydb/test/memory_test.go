package test

import (
	"testing"

	_ "github.com/go-kivik/memorydb/v4"
)

func init() {
	RegisterMemoryDBSuite()
}

func TestMemory(t *testing.T) {
	MemoryTest(t)
}
