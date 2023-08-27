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

package kiviktest

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"

	// Tests
	_ "github.com/go-kivik/kiviktest/v4/client"
	_ "github.com/go-kivik/kiviktest/v4/db"
)

// The available test suites
const (
	SuiteAuto        = "auto"
	SuitePouchLocal  = "pouch"
	SuitePouchRemote = "pouchRemote"
	SuiteCouch16     = "couch16"
	SuiteCouch17     = "couch17"
	SuiteCouch20     = "couch20"
	SuiteCouch21     = "couch21"
	SuiteCouch22     = "couch22"
	SuiteCouch23     = "couch23"
	SuiteCouch30     = "couch30"
	SuiteCouch31     = "couch31"
	SuiteCouch32     = "couch32"
	SuiteCouch33     = "couch33"
	SuiteCloudant    = "cloudant"
	SuiteKivikServer = "kivikServer"
	SuiteKivikMemory = "kivikMemory"
	SuiteKivikFS     = "kivikFilesystem"
)

// AllSuites is a list of all defined suites.
var AllSuites = []string{
	SuitePouchLocal,
	SuitePouchRemote,
	SuiteCouch16,
	SuiteCouch17,
	SuiteCouch20,
	SuiteCouch21,
	SuiteCouch22,
	SuiteCouch30,
	SuiteCouch31,
	SuiteCouch32,
	SuiteCouch33,
	SuiteKivikMemory,
	SuiteKivikFS,
	SuiteCloudant,
	SuiteKivikServer,
}

var driverMap = map[string]string{
	SuitePouchLocal:  "pouch",
	SuitePouchRemote: "pouch",
	SuiteCouch16:     "couch",
	SuiteCouch17:     "couch",
	SuiteCouch20:     "couch",
	SuiteCouch21:     "couch",
	SuiteCouch22:     "couch",
	SuiteCouch23:     "couch",
	SuiteCouch30:     "couch",
	SuiteCouch31:     "couch",
	SuiteCouch32:     "couch",
	SuiteCouch33:     "couch",
	SuiteCloudant:    "couch",
	SuiteKivikServer: "couch",
	SuiteKivikMemory: "memory",
	SuiteKivikFS:     "fs",
}

// ListTests prints a list of available test suites to stdout.
func ListTests() {
	fmt.Printf("Available test suites:\n\tauto\n")
	for _, suite := range AllSuites {
		fmt.Printf("\t%s\n", suite)
	}
}

// Options are the options to run a test from the command line tool.
type Options struct {
	Driver  string
	DSN     string
	Verbose bool
	RW      bool
	Match   string
	Suites  []string
	Cleanup bool
}

// CleanupTests attempts to clean up any stray test databases created by a
// previous test run.
func CleanupTests(driver, dsn string, verbose bool) error {
	client, err := kivik.New(driver, dsn)
	if err != nil {
		return err
	}
	count, err := doCleanup(client, verbose)
	if verbose {
		fmt.Printf("Deleted %d test databases\n", count)
	}
	return err
}

func doCleanup(client *kivik.Client, verbose bool) (int, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 3)
	var count int32
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		c, err := cleanupDatabases(ctx, client, verbose)
		if err != nil {
			cancel()
		}
		atomic.AddInt32(&count, int32(c))
		errCh <- err
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		c, err := cleanupUsers(ctx, client, verbose)
		if err != nil {
			cancel()
		}
		atomic.AddInt32(&count, int32(c))
		errCh <- err
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		c, err := cleanupReplications(ctx, client, verbose)
		if err != nil {
			cancel()
		}
		atomic.AddInt32(&count, int32(c))
		errCh <- err
	}()

	wg.Wait()
	err := <-errCh
	for len(errCh) > 0 {
		<-errCh
	}
	return int(count), err
}

func cleanupDatabases(ctx context.Context, client *kivik.Client, verbose bool) (int, error) {
	if verbose {
		fmt.Printf("Cleaning up stale databases\n")
	}
	allDBs, err := client.AllDBs(ctx)
	if err != nil {
		return 0, err
	}
	var count int
	for _, dbName := range allDBs {
		// FIXME: This filtering should be possible in AllDBs(), but all the
		// backends need to support it first.
		if strings.HasPrefix(dbName, kt.TestDBPrefix) {
			if verbose {
				fmt.Printf("\t--- Deleting %s\n", dbName)
			}
			if e := client.DestroyDB(ctx, dbName); e != nil && kivik.HTTPStatus(e) != http.StatusNotFound {
				return count, e
			}
			count++
		}
	}
	replicator := client.DB("_replicator")
	if e := replicator.Err(); e != nil {
		if kivik.HTTPStatus(e) != http.StatusNotFound && kivik.HTTPStatus(e) != http.StatusNotImplemented {
			return count, e
		}
		return count, nil
	}
	docs := replicator.AllDocs(context.Background(), map[string]interface{}{"include_docs": true})
	if err := docs.Err(); err != nil {
		if kivik.HTTPStatus(err) == http.StatusNotImplemented || kivik.HTTPStatus(err) == http.StatusNotFound {
			return count, nil
		}
		return count, err
	}
	var replDoc struct {
		Rev string `json:"_rev"`
	}
	for docs.Next() {
		id, _ := docs.ID()
		if strings.HasPrefix(id, "kivik$") {
			if err := docs.ScanDoc(&replDoc); err != nil {
				return count, err
			}
			if _, err := replicator.Delete(context.Background(), id, replDoc.Rev); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

func cleanupUsers(ctx context.Context, client *kivik.Client, verbose bool) (int, error) {
	if verbose {
		fmt.Printf("Cleaning up stale users\n")
	}
	db := client.DB("_users")
	if err := db.Err(); err != nil {
		switch kivik.HTTPStatus(err) {
		case http.StatusNotFound, http.StatusNotImplemented:
			return 0, nil
		}
		return 0, err
	}
	users := db.AllDocs(ctx, map[string]interface{}{"include_docs": true})
	if err := users.Err(); err != nil {
		switch kivik.HTTPStatus(err) {
		case http.StatusNotFound, http.StatusNotImplemented:
			return 0, nil
		}
		return 0, err
	}
	var count int
	for users.Next() {
		id, _ := users.ID()
		if strings.HasPrefix(id, "org.couchdb.user:kivik$") {
			if verbose {
				fmt.Printf("\t--- Deleting user %s\n", id)
			}
			var doc struct {
				Rev string `json:"_rev"`
			}
			if err := users.ScanDoc(&doc); err != nil {
				return count, err
			}
			if _, err := db.Delete(ctx, id, doc.Rev); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, users.Err()
}

func cleanupReplications(ctx context.Context, client *kivik.Client, verbose bool) (int, error) {
	if verbose {
		fmt.Printf("Cleaning up stale replications\n")
	}
	db := client.DB("_replicator")
	if err := db.Err(); err != nil {
		switch kivik.HTTPStatus(err) {
		case http.StatusNotFound, http.StatusNotImplemented:
			return 0, nil
		}
		return 0, err
	}
	reps := db.AllDocs(ctx, map[string]interface{}{"include_docs": true})
	if err := reps.Err(); err != nil {
		switch kivik.HTTPStatus(err) {
		case http.StatusNotFound, http.StatusNotImplemented:
			return 0, nil
		}
		return 0, err
	}
	var count int
	for reps.Next() {
		var doc struct {
			Rev    string `json:"_rev"`
			Source string `json:"source"`
			Target string `json:"target"`
		}
		if err := reps.ScanDoc(&doc); err != nil {
			return count, err
		}
		id, _ := reps.ID()
		if strings.HasPrefix(id, "kivik$") ||
			strings.HasPrefix(doc.Source, "kivik$") ||
			strings.HasPrefix(doc.Target, "kivik$") {
			if verbose {
				fmt.Printf("\t--- Deleting replication %s\n", id)
			}
			if _, err := db.Delete(ctx, id, doc.Rev); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, reps.Err()
}

// RunTests runs the requested test suites against the requested driver and DSN.
func RunTests(opts Options) {
	if opts.Cleanup {
		err := CleanupTests(opts.Driver, opts.DSN, opts.Verbose)
		if err != nil {
			fmt.Printf("Cleanup failed: %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	_ = flag.Set("test.run", opts.Match)
	if opts.Verbose {
		_ = flag.Set("test.v", "true")
	}
	tests := []testing.InternalTest{
		{
			Name: "MainTest",
			F: func(t *testing.T) {
				Test(t, opts.Driver, opts.DSN, opts.Suites, opts.RW)
			},
		},
	}

	mainStart(tests)
}

// Test is the main test entry point when running tests through the command line
// tool.
func Test(t *testing.T, driver, dsn string, testSuites []string, rw bool) {
	clients, err := ConnectClients(t, driver, dsn, nil)
	if err != nil {
		t.Fatalf("Failed to connect to %s (%s driver): %s\n", dsn, driver, err)
	}
	clients.RW = rw
	tests := make(map[string]struct{})
	for _, test := range testSuites {
		tests[test] = struct{}{}
	}
	if _, ok := tests[SuiteAuto]; ok {
		t.Log("Detecting target service compatibility...")
		suites, err := detectCompatibility(clients.Admin)
		if err != nil {
			t.Fatalf("Unable to determine server suite compatibility: %s\n", err)
		}
		tests = make(map[string]struct{})
		for _, suite := range suites {
			tests[suite] = struct{}{}
		}
	}
	testSuites = make([]string, 0, len(tests))
	for test := range tests {
		testSuites = append(testSuites, test)
	}
	t.Logf("Running the following test suites: %s\n", strings.Join(testSuites, ", "))
	for _, suite := range testSuites {
		RunTestsInternal(clients, suite)
	}
}

// RunTestsInternal is for internal use only.
func RunTestsInternal(ctx *kt.Context, suite string) {
	conf, ok := suites[suite]
	if !ok {
		ctx.Skipf("No configuration found for suite '%s'", suite)
	}
	ctx.Config = conf
	// This is run as a sub-test so configuration will work nicely.
	ctx.Run("PreCleanup", func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			count, err := doCleanup(ctx.Admin, true)
			if count > 0 {
				ctx.Logf("Pre-cleanup removed %d databases from previous test runs", count)
			}
			if err != nil {
				ctx.Fatalf("Pre-cleanup failed: %s", err)
			}
		})
	})
	kt.RunSubtests(ctx)
}

func detectCompatibility(client *kivik.Client) ([]string, error) {
	info, err := client.Version(context.Background())
	if err != nil {
		return nil, err
	}
	switch info.Vendor {
	case "PouchDB":
		return []string{SuitePouchLocal}, nil
	case "IBM Cloudant":
		return []string{SuiteCloudant}, nil
	case "The Apache Software Foundation":
		if strings.HasPrefix(info.Version, "2.0") {
			return []string{SuiteCouch20}, nil
		}
		if strings.HasPrefix(info.Version, "2.1") {
			return []string{SuiteCouch21}, nil
		}
		return []string{SuiteCouch16}, nil
	case "Kivik Memory Adaptor":
		return []string{SuiteKivikMemory}, nil
	}
	return []string{}, errors.New("Unable to automatically determine the proper test suite")
}

// ConnectClients connects clients.
func ConnectClients(t *testing.T, driverName, dsn string, opts kivik.Options) (*kt.Context, error) {
	var noAuthDSN string
	if parsed, err := url.Parse(dsn); err == nil {
		if parsed.User == nil {
			return nil, errors.New("DSN does not contain authentication credentials")
		}
		parsed.User = nil
		noAuthDSN = parsed.String()
	}
	clients := &kt.Context{
		T: t,
	}
	t.Logf("Connecting to %s ...\n", dsn)
	if client, err := kivik.New(driverName, dsn, opts); err == nil {
		clients.Admin = client
	} else {
		return nil, err
	}

	t.Logf("Connecting to %s ...\n", noAuthDSN)
	if client, err := kivik.New(driverName, noAuthDSN, opts); err == nil {
		clients.NoAuth = client
	} else {
		return nil, err
	}
	return clients, nil
}

// DoTest runs a suite of tests.
func DoTest(t *testing.T, suite, envName string) {
	opts, _ := suites[suite].Interface(t, "Options").(kivik.Options)

	dsn := os.Getenv(envName)
	if dsn == "" {
		t.Skipf("%s: %s DSN not set; skipping tests", envName, suite)
	}
	clients, err := ConnectClients(t, driverMap[suite], dsn, opts)
	if err != nil {
		t.Errorf("Failed to connect to %s: %s\n", suite, err)
		return
	}
	clients.RW = true
	RunTestsInternal(clients, suite)
}
