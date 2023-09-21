//go:build go1.8
// +build go1.8

package memorydb

import "fmt"

func randStr() string {
	rndMU.Lock()
	s := fmt.Sprintf("%016x%016x", rnd.Uint64(), rnd.Uint64())
	rndMU.Unlock()
	return s
}
