// Package main prints out the version constant, for use in automatically
// creating releases when the version is updated.
package main

import (
	"fmt"
	"strings"

	"github.com/go-kivik/kivik/v4"
)

func main() {
	if strings.HasSuffix(kivik.Version, "-prerelease") {
		return
	}
	fmt.Printf("TAG=%s\n", kivik.Version)
	if strings.Contains(kivik.Version, "-") {
		fmt.Println("PRERELEASE=true")
	}
}
