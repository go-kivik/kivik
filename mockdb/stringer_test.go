package kivikmock

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
)

type stringerTest struct {
	input    fmt.Stringer
	expected string
}

func testStringer(t *testing.T, test stringerTest) {
	t.Helper()
	result := test.input.String()
	if test.expected != result {
		t.Errorf("Unexpected String() output.\nWant: %s\n Got: %s\n", test.expected, result)
	}
}

func TestCloseString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("standard", stringerTest{
		input:    &ExpectedClose{},
		expected: "call to Close()",
	})
	tests.Add("error", stringerTest{
		input: &ExpectedClose{commonExpectation: commonExpectation{err: errors.New("foo error")}},
		expected: `call to Close() which:
	- should return error: foo error`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedClose{commonExpectation: commonExpectation{delay: time.Second}},
		expected: `call to Close() which:
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestReplicateString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("standard", stringerTest{
		input: &ExpectedReplicate{},
		expected: `call to Replicate() which:
	- has any target
	- has any source
	- has any options`,
	})
	tests.Add("target", stringerTest{
		input: &ExpectedReplicate{arg0: "foo"},
		expected: `call to Replicate() which:
	- has target: foo
	- has any source
	- has any options`,
	})
	tests.Add("source", stringerTest{
		input: &ExpectedReplicate{arg1: "http://example.com/bar"},
		expected: `call to Replicate() which:
	- has any target
	- has source: http://example.com/bar
	- has any options`,
	})
	tests.Add("return", stringerTest{
		input: &ExpectedReplicate{ret0: &Replication{id: "foo"}},
		expected: `call to Replicate() which:
	- has any target
	- has any source
	- has any options
	- should return: {"replication_id":"foo"}`,
	})
	tests.Add("options", stringerTest{
		input: &ExpectedReplicate{commonExpectation: commonExpectation{options: map[string]interface{}{"foo": 123}}},
		expected: `call to Replicate() which:
	- has any target
	- has any source
	- has options: map[foo:123]`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedReplicate{commonExpectation: commonExpectation{err: errors.New("foo error")}},
		expected: `call to Replicate() which:
	- has any target
	- has any source
	- has any options
	- should return error: foo error`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedReplicate{commonExpectation: commonExpectation{delay: time.Second}},
		expected: `call to Replicate() which:
	- has any target
	- has any source
	- has any options
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestGetReplicationsString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("standard", stringerTest{
		input: &ExpectedGetReplications{},
		expected: `call to GetReplications() which:
	- has any options`,
	})
	tests.Add("options", stringerTest{
		input: &ExpectedGetReplications{commonExpectation: commonExpectation{options: map[string]interface{}{"foo": 123}}},
		expected: `call to GetReplications() which:
	- has options: map[foo:123]`,
	})
	tests.Add("return", stringerTest{
		input: &ExpectedGetReplications{ret0: []*Replication{{}, {}}},
		expected: `call to GetReplications() which:
	- has any options
	- should return: 2 results`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedGetReplications{commonExpectation: commonExpectation{err: errors.New("foo error")}},
		expected: `call to GetReplications() which:
	- has any options
	- should return error: foo error`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedGetReplications{commonExpectation: commonExpectation{delay: time.Second}},
		expected: `call to GetReplications() which:
	- has any options
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestAllDBsString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("standard", stringerTest{
		input: &ExpectedAllDBs{},
		expected: `call to AllDBs() which:
	- has any options`,
	})
	tests.Add("options", stringerTest{
		input: &ExpectedAllDBs{commonExpectation: commonExpectation{options: map[string]interface{}{"foo": 123}}},
		expected: `call to AllDBs() which:
	- has options: map[foo:123]`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedAllDBs{commonExpectation: commonExpectation{err: errors.New("foo err")}},
		expected: `call to AllDBs() which:
	- has any options
	- should return error: foo err`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedAllDBs{commonExpectation: commonExpectation{delay: time.Second}},
		expected: `call to AllDBs() which:
	- has any options
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestAuthenticateString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedAuthenticate{},
		expected: `call to Authenticate() which:
	- has any authenticator`,
	})
	tests.Add("authenticator", stringerTest{
		input: &ExpectedAuthenticate{authType: "foo"},
		expected: `call to Authenticate() which:
	- has an authenticator of type: foo`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedAuthenticate{commonExpectation: commonExpectation{err: errors.New("foo err")}},
		expected: `call to Authenticate() which:
	- has any authenticator
	- should return error: foo err`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedAuthenticate{commonExpectation: commonExpectation{delay: time.Second}},
		expected: `call to Authenticate() which:
	- has any authenticator
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestClusterSetupString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedClusterSetup{},
		expected: `call to ClusterSetup() which:
	- has any action`,
	})
	tests.Add("action", stringerTest{
		input: &ExpectedClusterSetup{arg0: map[string]string{"foo": "bar"}},
		expected: `call to ClusterSetup() which:
	- has the action: map[foo:bar]`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedClusterSetup{commonExpectation: commonExpectation{err: errors.New("foo error")}},
		expected: `call to ClusterSetup() which:
	- has any action
	- should return error: foo error`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedClusterSetup{commonExpectation: commonExpectation{delay: time.Second}},
		expected: `call to ClusterSetup() which:
	- has any action
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestClusterStatusString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedClusterStatus{},
		expected: `call to ClusterStatus() which:
	- has any options`,
	})
	tests.Add("options", stringerTest{
		input: &ExpectedClusterStatus{commonExpectation: commonExpectation{options: map[string]interface{}{"foo": 123}}},
		expected: `call to ClusterStatus() which:
	- has options: map[foo:123]`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedClusterStatus{commonExpectation: commonExpectation{err: errors.New("foo error")}},
		expected: `call to ClusterStatus() which:
	- has any options
	- should return error: foo error`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedClusterStatus{commonExpectation: commonExpectation{delay: time.Second}},
		expected: `call to ClusterStatus() which:
	- has any options
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestMembershipString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input:    &ExpectedMembership{},
		expected: `call to Membership()`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedMembership{commonExpectation: commonExpectation{err: errors.New("foo error")}},
		expected: `call to Membership() which:
	- should return error: foo error`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedMembership{commonExpectation: commonExpectation{delay: time.Second}},
		expected: `call to Membership() which:
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestDBExistsString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedDBExists{},
		expected: `call to DBExists() which:
	- has any name
	- has any options
	- should return: false`,
	})
	tests.Add("full", stringerTest{
		input: &ExpectedDBExists{arg0: "foo", ret0: true, commonExpectation: commonExpectation{options: map[string]interface{}{"foo": 123}}},
		expected: `call to DBExists() which:
	- has name: foo
	- has options: map[foo:123]
	- should return: true`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedDBExists{commonExpectation: commonExpectation{err: errors.New("foo err")}},
		expected: `call to DBExists() which:
	- has any name
	- has any options
	- should return error: foo err`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedDBExists{commonExpectation: commonExpectation{delay: time.Second}},
		expected: `call to DBExists() which:
	- has any name
	- has any options
	- should delay for: 1s
	- should return: false`,
	})
	tests.Run(t, testStringer)
}

func TestDestroyDBString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedDestroyDB{},
		expected: `call to DestroyDB() which:
	- has any name
	- has any options`,
	})
	tests.Add("name", stringerTest{
		input: &ExpectedDestroyDB{arg0: "foo"},
		expected: `call to DestroyDB() which:
	- has name: foo
	- has any options`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedDestroyDB{commonExpectation: commonExpectation{delay: time.Second}},
		expected: `call to DestroyDB() which:
	- has any name
	- has any options
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestDBsStatsString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedDBsStats{},
		expected: `call to DBsStats() which:
	- has any names`,
	})
	tests.Add("names", stringerTest{
		input: &ExpectedDBsStats{arg0: []string{"a", "b"}},
		expected: `call to DBsStats() which:
	- has names: [a b]`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedDBsStats{commonExpectation: commonExpectation{delay: time.Second}},
		expected: `call to DBsStats() which:
	- has any names
	- should delay for: 1s`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedDBsStats{commonExpectation: commonExpectation{err: errors.New("foo err")}},
		expected: `call to DBsStats() which:
	- has any names
	- should return error: foo err`,
	})
	tests.Run(t, testStringer)
}

func TestPingString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input:    &ExpectedPing{},
		expected: `call to Ping()`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedPing{commonExpectation: commonExpectation{err: errors.New("foo err")}},
		expected: `call to Ping() which:
	- should return error: foo err`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedPing{commonExpectation: commonExpectation{delay: time.Second}},
		expected: `call to Ping() which:
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestSessionString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input:    &ExpectedSession{},
		expected: `call to Session()`,
	})
	tests.Add("session", stringerTest{
		input: &ExpectedSession{ret0: &driver.Session{Name: "bob"}},
		expected: `call to Session() which:
	- should return: &{bob []   [] []}`,
	})
	tests.Run(t, testStringer)
}

func TestVersionString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input:    &ExpectedVersion{},
		expected: `call to Version()`,
	})
	tests.Add("session", stringerTest{
		input: &ExpectedVersion{ret0: &driver.Version{Version: "1.2"}},
		expected: `call to Version() which:
	- should return: &{1.2  [] []}`,
	})
	tests.Run(t, testStringer)
}

func TestCreateDBString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedCreateDB{},
		expected: `call to CreateDB() which:
	- has any name
	- has any options`,
	})
	tests.Add("name", stringerTest{
		input: &ExpectedCreateDB{arg0: "foo"},
		expected: `call to CreateDB() which:
	- has name: foo
	- has any options`,
	})
	tests.Add("db", stringerTest{
		input: &ExpectedCreateDB{commonExpectation: commonExpectation{db: &DB{count: 50}}},
		expected: `call to CreateDB() which:
	- has any name
	- has any options
	- should return database with 50 expectations`,
	})
	tests.Run(t, testStringer)
}

func TestDBString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedDB{},
		expected: `call to DB() which:
	- has any name
	- has any options`,
	})
	tests.Add("name", stringerTest{
		input: &ExpectedDB{arg0: "foo"},
		expected: `call to DB() which:
	- has name: foo
	- has any options`,
	})
	tests.Add("db", stringerTest{
		input: &ExpectedDB{commonExpectation: commonExpectation{db: &DB{count: 50}}},
		expected: `call to DB() which:
	- has any name
	- has any options
	- should return database with 50 expectations`,
	})
	tests.Run(t, testStringer)
}

func TestDBCloseString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("standard", stringerTest{
		input:    &ExpectedDBClose{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: "call to DB(foo#0).Close()",
	})
	tests.Add("error", stringerTest{
		input: &ExpectedDBClose{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo error")}},
		expected: `call to DB(foo#0).Close() which:
	- should return error: foo error`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedDBClose{commonExpectation: commonExpectation{db: &DB{name: "foo"}, delay: time.Second}},
		expected: `call to DB(foo#0).Close() which:
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestAllDocsString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedAllDocs{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).AllDocs() which:
	- has any options`,
	})
	tests.Add("results", stringerTest{
		input: &ExpectedAllDocs{
			commonExpectation: commonExpectation{db: &DB{name: "foo"}},
			ret0: &Rows{iter: iter{items: []*item{
				{item: &driver.Row{}},
				{item: &driver.Row{}},
				{delay: 15},
				{item: &driver.Row{}},
				{item: &driver.Row{}},
			}}},
		},
		expected: `call to DB(foo#0).AllDocs() which:
	- has any options
	- should return: 4 results`,
	})
	tests.Run(t, testStringer)
}

func TestBulkGetString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedBulkGet{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).BulkGet() which:
	- has any doc references
	- has any options`,
	})
	tests.Add("docs", stringerTest{
		input: &ExpectedBulkGet{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: []driver.BulkGetReference{
			{ID: "foo"},
			{ID: "bar"},
		}},
		expected: `call to DB(foo#0).BulkGet() which:
	- has doc references: [{"id":"foo"},{"id":"bar"}]
	- has any options`,
	})
	tests.Add("results", stringerTest{
		input: &ExpectedBulkGet{
			commonExpectation: commonExpectation{db: &DB{name: "foo"}},
			ret0: &Rows{iter: iter{items: []*item{
				{item: &driver.Row{}},
				{item: &driver.Row{}},
				{delay: 15},
				{item: &driver.Row{}},
				{item: &driver.Row{}},
			}}},
		},
		expected: `call to DB(foo#0).BulkGet() which:
	- has any doc references
	- has any options
	- should return: 4 results`,
	})
	tests.Run(t, testStringer)
}

func TestFindString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedFind{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Find() which:
	- has any query
	- has any options`,
	})
	tests.Add("query", stringerTest{
		input: &ExpectedFind{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: map[string]string{"foo": "bar"}},
		expected: `call to DB(foo#0).Find() which:
	- has query: map[foo:bar]
	- has any options`,
	})
	tests.Add("results", stringerTest{
		input: &ExpectedFind{
			commonExpectation: commonExpectation{db: &DB{name: "foo"}},
			ret0: &Rows{iter: iter{items: []*item{
				{item: &driver.Row{}},
				{item: &driver.Row{}},
				{delay: 15},
				{item: &driver.Row{}},
				{item: &driver.Row{}},
			}}},
		},
		expected: `call to DB(foo#0).Find() which:
	- has any query
	- has any options
	- should return: 4 results`,
	})
	tests.Run(t, testStringer)
}

func TestCreateIndexString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedCreateIndex{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).CreateIndex() which:
	- has any ddoc
	- has any name
	- has any index`,
	})
	tests.Add("ddoc", stringerTest{
		input: &ExpectedCreateIndex{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).CreateIndex() which:
	- has ddoc: foo
	- has any name
	- has any index`,
	})
	tests.Add("name", stringerTest{
		input: &ExpectedCreateIndex{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg1: "foo"},
		expected: `call to DB(foo#0).CreateIndex() which:
	- has any ddoc
	- has name: foo
	- has any index`,
	})
	tests.Add("index", stringerTest{
		input: &ExpectedCreateIndex{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg2: map[string]string{"foo": "bar"}},
		expected: `call to DB(foo#0).CreateIndex() which:
	- has any ddoc
	- has any name
	- has index: map[foo:bar]`,
	})
	tests.Run(t, testStringer)
}

func TestExpectedGetIndexesString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input:    &ExpectedGetIndexes{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).GetIndexes()`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedGetIndexes{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo err")}},
		expected: `call to DB(foo#0).GetIndexes() which:
	- should return error: foo err`,
	})
	tests.Add("indexes", stringerTest{
		input: &ExpectedGetIndexes{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: []driver.Index{{Name: "foo"}}},
		expected: `call to DB(foo#0).GetIndexes() which:
	- should return indexes: [{ foo  <nil>}]`,
	})
	tests.Run(t, testStringer)
}

func TestDeleteIndexString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedDeleteIndex{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).DeleteIndex() which:
	- has any ddoc
	- has any name`,
	})
	tests.Add("ddoc", stringerTest{
		input: &ExpectedDeleteIndex{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).DeleteIndex() which:
	- has ddoc: foo
	- has any name`,
	})
	tests.Add("name", stringerTest{
		input: &ExpectedDeleteIndex{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg1: "foo"},
		expected: `call to DB(foo#0).DeleteIndex() which:
	- has any ddoc
	- has name: foo`,
	})
	tests.Run(t, testStringer)
}

func TestExplainString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedExplain{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Explain() which:
	- has any query`,
	})
	tests.Add("query", stringerTest{
		input: &ExpectedExplain{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: map[string]string{"foo": "bar"}},
		expected: `call to DB(foo#0).Explain() which:
	- has query: map[foo:bar]`,
	})
	tests.Add("plan", stringerTest{
		input: &ExpectedExplain{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: &driver.QueryPlan{DBName: "foo"}},
		expected: `call to DB(foo#0).Explain() which:
	- has any query
	- should return query plan: &{foo map[] map[] map[] 0 0 [] map[]}`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedExplain{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo err")}},
		expected: `call to DB(foo#0).Explain() which:
	- has any query
	- should return error: foo err`,
	})
	tests.Run(t, testStringer)
}

func TestCreateDocString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedCreateDoc{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).CreateDoc() which:
	- has any doc
	- has any options`,
	})
	tests.Add("doc", stringerTest{
		input: &ExpectedCreateDoc{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: map[string]string{"foo": "bar"}},
		expected: `call to DB(foo#0).CreateDoc() which:
	- has doc: {"foo":"bar"}
	- has any options`,
	})
	tests.Add("options", stringerTest{
		input: &ExpectedCreateDoc{commonExpectation: commonExpectation{db: &DB{name: "foo"}, options: map[string]interface{}{"foo": "bar"}}},
		expected: `call to DB(foo#0).CreateDoc() which:
	- has any doc
	- has options: map[foo:bar]`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedCreateDoc{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo err")}},
		expected: `call to DB(foo#0).CreateDoc() which:
	- has any doc
	- has any options
	- should return error: foo err`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedCreateDoc{commonExpectation: commonExpectation{db: &DB{name: "foo"}, delay: time.Second}},
		expected: `call to DB(foo#0).CreateDoc() which:
	- has any doc
	- has any options
	- should delay for: 1s`,
	})
	tests.Add("docID", stringerTest{
		input: &ExpectedCreateDoc{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: "foo"},
		expected: `call to DB(foo#0).CreateDoc() which:
	- has any doc
	- has any options
	- should return docID: foo`,
	})
	tests.Add("rev", stringerTest{
		input: &ExpectedCreateDoc{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret1: "1-foo"},
		expected: `call to DB(foo#0).CreateDoc() which:
	- has any doc
	- has any options
	- should return rev: 1-foo`,
	})
	tests.Run(t, testStringer)
}

func TestCompactString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input:    &ExpectedCompact{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Compact()`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedCompact{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo err")}},
		expected: `call to DB(foo#0).Compact() which:
	- should return error: foo err`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedCompact{commonExpectation: commonExpectation{db: &DB{name: "foo"}, delay: time.Second}},
		expected: `call to DB(foo#0).Compact() which:
	- should delay for: 1s`,
	})

	tests.Run(t, testStringer)
}

func TestViewCleanupString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input:    &ExpectedViewCleanup{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).ViewCleanup()`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedViewCleanup{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo err")}},
		expected: `call to DB(foo#0).ViewCleanup() which:
	- should return error: foo err`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedViewCleanup{commonExpectation: commonExpectation{db: &DB{name: "foo"}, delay: time.Second}},
		expected: `call to DB(foo#0).ViewCleanup() which:
	- should delay for: 1s`,
	})

	tests.Run(t, testStringer)
}

func TestPutString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedPut{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Put() which:
	- has any docID
	- has any doc
	- has any options`,
	})
	tests.Add("docID", stringerTest{
		input: &ExpectedPut{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).Put() which:
	- has docID: foo
	- has any doc
	- has any options`,
	})
	tests.Add("doc", stringerTest{
		input: &ExpectedPut{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg1: map[string]string{"foo": "bar"}},
		expected: `call to DB(foo#0).Put() which:
	- has any docID
	- has doc: {"foo":"bar"}
	- has any options`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedPut{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo err")}},
		expected: `call to DB(foo#0).Put() which:
	- has any docID
	- has any doc
	- has any options
	- should return error: foo err`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedPut{commonExpectation: commonExpectation{db: &DB{name: "foo"}, delay: time.Second}},
		expected: `call to DB(foo#0).Put() which:
	- has any docID
	- has any doc
	- has any options
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestGetRevString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedGetRev{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).GetRev() which:
	- has any docID
	- has any options`,
	})
	tests.Add("docID", stringerTest{
		input: &ExpectedGetRev{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).GetRev() which:
	- has docID: foo
	- has any options`,
	})
	tests.Add("rev", stringerTest{
		input: &ExpectedGetRev{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: "1-xxx"},
		expected: `call to DB(foo#0).GetRev() which:
	- has any docID
	- has any options
	- should return rev: 1-xxx`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedGetRev{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo err")}},
		expected: `call to DB(foo#0).GetRev() which:
	- has any docID
	- has any options
	- should return error: foo err`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedGetRev{commonExpectation: commonExpectation{db: &DB{name: "foo"}, delay: time.Second}},
		expected: `call to DB(foo#0).GetRev() which:
	- has any docID
	- has any options
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestCompactViewString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedCompactView{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).CompactView() which:
	- has any ddocID`,
	})
	tests.Add("ddocID", stringerTest{
		input: &ExpectedCompactView{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).CompactView() which:
	- has ddocID: foo`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedCompactView{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo err")}},
		expected: `call to DB(foo#0).CompactView() which:
	- has any ddocID
	- should return error: foo err`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedCompactView{commonExpectation: commonExpectation{db: &DB{name: "foo"}, delay: time.Second}},
		expected: `call to DB(foo#0).CompactView() which:
	- has any ddocID
	- should delay for: 1s`,
	})
	tests.Run(t, testStringer)
}

func TestFlushString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input:    &ExpectedFlush{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Flush()`,
	})
	tests.Run(t, testStringer)
}

func TestDeleteAttachmentString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedDeleteAttachment{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).DeleteAttachment() which:
	- has any docID
	- has any filename
	- has any options`,
	})
	tests.Add("docID", stringerTest{
		input: &ExpectedDeleteAttachment{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).DeleteAttachment() which:
	- has docID: foo
	- has any filename
	- has any options`,
	})
	tests.Add("filename", stringerTest{
		input: &ExpectedDeleteAttachment{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg1: "foo.txt"},
		expected: `call to DB(foo#0).DeleteAttachment() which:
	- has any docID
	- has filename: foo.txt
	- has any options`,
	})
	tests.Add("return", stringerTest{
		input: &ExpectedDeleteAttachment{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: "2-bar"},
		expected: `call to DB(foo#0).DeleteAttachment() which:
	- has any docID
	- has any filename
	- has any options
	- should return rev: 2-bar`,
	})
	tests.Run(t, testStringer)
}

func TestDeleteString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedDelete{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Delete() which:
	- has any docID
	- has any options`,
	})
	tests.Add("docID", stringerTest{
		input: &ExpectedDelete{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).Delete() which:
	- has docID: foo
	- has any options`,
	})
	tests.Add("options", stringerTest{
		input: &ExpectedDelete{commonExpectation: commonExpectation{db: &DB{name: "foo"}, options: map[string]interface{}{"foo": "bar"}}},
		expected: `call to DB(foo#0).Delete() which:
	- has any docID
	- has options: map[foo:bar]`,
	})
	tests.Add("return", stringerTest{
		input: &ExpectedDelete{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: "2-bar"},
		expected: `call to DB(foo#0).Delete() which:
	- has any docID
	- has any options
	- should return rev: 2-bar`,
	})
	tests.Run(t, testStringer)
}

func TestCopyString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedCopy{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Copy() which:
	- has any targetID
	- has any sourceID
	- has any options`,
	})
	tests.Add("targetID", stringerTest{
		input: &ExpectedCopy{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).Copy() which:
	- has targetID: foo
	- has any sourceID
	- has any options`,
	})
	tests.Add("sourceID", stringerTest{
		input: &ExpectedCopy{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg1: "foo"},
		expected: `call to DB(foo#0).Copy() which:
	- has any targetID
	- has sourceID: foo
	- has any options`,
	})
	tests.Add("return value", stringerTest{
		input: &ExpectedCopy{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: "1-foo"},
		expected: `call to DB(foo#0).Copy() which:
	- has any targetID
	- has any sourceID
	- has any options
	- should return rev: 1-foo`,
	})
	tests.Run(t, testStringer)
}

func TestGetString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedGet{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Get() which:
	- has any docID
	- has any options`,
	})
	tests.Add("docID", stringerTest{
		input: &ExpectedGet{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).Get() which:
	- has docID: foo
	- has any options`,
	})
	tests.Add("return value", stringerTest{
		input: &ExpectedGet{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: &driver.Document{Rev: "1-foo"}},
		expected: `call to DB(foo#0).Get() which:
	- has any docID
	- has any options
	- should return document with rev: 1-foo`,
	})
	tests.Run(t, testStringer)
}

func TestGetAttachmentMetaString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedGetAttachmentMeta{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).GetAttachmentMeta() which:
	- has any docID
	- has any filename
	- has any options`,
	})
	tests.Add("docID", stringerTest{
		input: &ExpectedGetAttachmentMeta{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).GetAttachmentMeta() which:
	- has docID: foo
	- has any filename
	- has any options`,
	})
	tests.Add("filename", stringerTest{
		input: &ExpectedGetAttachmentMeta{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg1: "foo.txt"},
		expected: `call to DB(foo#0).GetAttachmentMeta() which:
	- has any docID
	- has filename: foo.txt
	- has any options`,
	})
	tests.Add("return value", stringerTest{
		input: &ExpectedGetAttachmentMeta{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: &driver.Attachment{Filename: "foo.txt"}},
		expected: `call to DB(foo#0).GetAttachmentMeta() which:
	- has any docID
	- has any filename
	- has any options
	- should return attachment: foo.txt`,
	})
	tests.Run(t, testStringer)
}

func TestLocalDocsString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedLocalDocs{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).LocalDocs() which:
	- has any options`,
	})
	tests.Add("results", stringerTest{
		input: &ExpectedLocalDocs{
			commonExpectation: commonExpectation{db: &DB{name: "foo"}},
			ret0: &Rows{iter: iter{items: []*item{
				{item: &driver.Row{}},
				{item: &driver.Row{}},
				{delay: 15},
				{item: &driver.Row{}},
				{item: &driver.Row{}},
			}}},
		},
		expected: `call to DB(foo#0).LocalDocs() which:
	- has any options
	- should return: 4 results`,
	})
	tests.Run(t, testStringer)
}

func TestPurgeString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedPurge{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Purge() which:
	- has any docRevMap`,
	})
	tests.Add("docRevMap", stringerTest{
		input: &ExpectedPurge{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: map[string][]string{"foo": {"a", "b"}}},
		expected: `call to DB(foo#0).Purge() which:
	- has docRevMap: map[foo:[a b]]`,
	})
	tests.Add("return", stringerTest{
		input: &ExpectedPurge{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: &driver.PurgeResult{Seq: 123}},
		expected: `call to DB(foo#0).Purge() which:
	- has any docRevMap
	- should return result: &{123 map[]}`,
	})
	tests.Run(t, testStringer)
}

func TestPutAttachmentString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedPutAttachment{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).PutAttachment() which:
	- has any docID
	- has any attachment
	- has any options`,
	})
	tests.Add("docID", stringerTest{
		input: &ExpectedPutAttachment{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).PutAttachment() which:
	- has docID: foo
	- has any attachment
	- has any options`,
	})
	tests.Add("attachment", stringerTest{
		input: &ExpectedPutAttachment{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg1: &driver.Attachment{Filename: "foo.txt"}},
		expected: `call to DB(foo#0).PutAttachment() which:
	- has any docID
	- has attachment: foo.txt
	- has any options`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedPutAttachment{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo err")}},
		expected: `call to DB(foo#0).PutAttachment() which:
	- has any docID
	- has any attachment
	- has any options
	- should return error: foo err`,
	})
	tests.Add("return value", stringerTest{
		input: &ExpectedPutAttachment{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: "2-bar"},
		expected: `call to DB(foo#0).PutAttachment() which:
	- has any docID
	- has any attachment
	- has any options
	- should return rev: 2-bar`,
	})
	tests.Run(t, testStringer)
}

func TestQueryString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedQuery{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Query() which:
	- has any ddocID
	- has any view
	- has any options`,
	})
	tests.Add("docID", stringerTest{
		input: &ExpectedQuery{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).Query() which:
	- has ddocID: foo
	- has any view
	- has any options`,
	})
	tests.Add("view", stringerTest{
		input: &ExpectedQuery{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg1: "1-foo"},
		expected: `call to DB(foo#0).Query() which:
	- has any ddocID
	- has view: 1-foo
	- has any options`,
	})
	tests.Add("results", stringerTest{
		input: &ExpectedQuery{
			commonExpectation: commonExpectation{db: &DB{name: "foo"}},
			ret0: &Rows{iter: iter{items: []*item{
				{item: &driver.Row{}},
				{item: &driver.Row{}},
				{delay: 15},
				{item: &driver.Row{}},
				{item: &driver.Row{}},
			}}},
		},
		expected: `call to DB(foo#0).Query() which:
	- has any ddocID
	- has any view
	- has any options
	- should return: 4 results`,
	})
	tests.Run(t, testStringer)
}

func TestGetAttachmentString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedGetAttachment{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).GetAttachment() which:
	- has any docID
	- has any filename
	- has any options`,
	})
	tests.Add("docID", stringerTest{
		input: &ExpectedGetAttachment{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).GetAttachment() which:
	- has docID: foo
	- has any filename
	- has any options`,
	})
	tests.Add("filename", stringerTest{
		input: &ExpectedGetAttachment{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg1: "foo.txt"},
		expected: `call to DB(foo#0).GetAttachment() which:
	- has any docID
	- has filename: foo.txt
	- has any options`,
	})
	tests.Add("return value", stringerTest{
		input: &ExpectedGetAttachment{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: &driver.Attachment{Filename: "foo.txt"}},
		expected: `call to DB(foo#0).GetAttachment() which:
	- has any docID
	- has any filename
	- has any options
	- should return attachment: foo.txt`,
	})
	tests.Run(t, testStringer)
}

func TestStatsString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input:    &ExpectedStats{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Stats()`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedStats{commonExpectation: commonExpectation{db: &DB{name: "foo"}, delay: time.Second}},
		expected: `call to DB(foo#0).Stats() which:
	- should delay for: 1s`,
	})
	tests.Add("return value", stringerTest{
		input: &ExpectedStats{ret0: &driver.DBStats{Name: "foo"}, commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Stats() which:
	- should return stats: &{foo false 0 0  0 0 0 <nil> []}`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedStats{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo err")}},
		expected: `call to DB(foo#0).Stats() which:
	- should return error: foo err`,
	})
	tests.Run(t, testStringer)
}

func TestBulkDocsString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedBulkDocs{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).BulkDocs() which:
	- has any docs
	- has any options`,
	})
	tests.Add("docs", stringerTest{
		input: &ExpectedBulkDocs{arg0: []interface{}{1, 2, 3}, commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).BulkDocs() which:
	- has: 3 docs
	- has any options`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedBulkDocs{commonExpectation: commonExpectation{db: &DB{name: "foo"}, delay: time.Second}},
		expected: `call to DB(foo#0).BulkDocs() which:
	- has any docs
	- has any options
	- should delay for: 1s`,
	})
	tests.Add("return value", stringerTest{
		input: &ExpectedBulkDocs{
			commonExpectation: commonExpectation{db: &DB{name: "foo"}},
			ret0: []driver.BulkResult{
				{},
				{},
				{},
			},
		},
		expected: `call to DB(foo#0).BulkDocs() which:
	- has any docs
	- has any options
	- should return: 3 results`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedBulkDocs{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo err")}},
		expected: `call to DB(foo#0).BulkDocs() which:
	- has any docs
	- has any options
	- should return error: foo err`,
	})
	tests.Run(t, testStringer)
}

func TestChangesString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedChanges{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Changes() which:
	- has any options`,
	})
	tests.Add("results", stringerTest{
		input: &ExpectedChanges{
			commonExpectation: commonExpectation{db: &DB{name: "foo"}},
			ret0: &Changes{iter: iter{items: []*item{
				{item: &driver.Change{}},
				{item: &driver.Change{}},
				{delay: 15},
				{item: &driver.Change{}},
				{item: &driver.Change{}},
			}}},
		},
		expected: `call to DB(foo#0).Changes() which:
	- has any options
	- should return: 4 results`,
	})
	tests.Run(t, testStringer)
}

func TestDBUpdatesString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input:    &ExpectedDBUpdates{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DBUpdates()`,
	})
	tests.Add("results", stringerTest{
		input: &ExpectedDBUpdates{
			commonExpectation: commonExpectation{db: &DB{name: "foo"}},
			ret0: &Updates{iter: iter{items: []*item{
				{item: &driver.DBUpdate{}},
				{item: &driver.DBUpdate{}},
				{delay: 15},
				{item: &driver.DBUpdate{}},
				{item: &driver.DBUpdate{}},
			}}},
		},
		expected: `call to DBUpdates() which:
	- should return: 4 results`,
	})
	tests.Run(t, testStringer)
}

func TestConfigString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedConfig{},
		expected: `call to Config() which:
	- has any node`,
	})
	tests.Add("node", stringerTest{
		input: &ExpectedConfig{arg0: "local"},
		expected: `call to Config() which:
	- has node: local`,
	})
	tests.Add("results", stringerTest{
		input: &ExpectedConfig{ret0: driver.Config{"foo": driver.ConfigSection{"bar": "baz"}}},
		expected: `call to Config() which:
	- has any node
	- should return: map[foo:map[bar:baz]]`,
	})

	tests.Run(t, testStringer)
}

func TestConfigSectionString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedConfigSection{},
		expected: `call to ConfigSection() which:
	- has any node
	- has any section`,
	})
	tests.Add("full", stringerTest{
		input: &ExpectedConfigSection{arg0: "local", arg1: "httpd"},
		expected: `call to ConfigSection() which:
	- has node: local
	- has section: httpd`,
	})
	tests.Add("results", stringerTest{
		input: &ExpectedConfigSection{ret0: driver.ConfigSection{"bar": "baz"}},
		expected: `call to ConfigSection() which:
	- has any node
	- has any section
	- should return: map[bar:baz]`,
	})

	tests.Run(t, testStringer)
}

func TestConfigValueString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedConfigValue{},
		expected: `call to ConfigValue() which:
	- has any node
	- has any section
	- has any key`,
	})
	tests.Add("full", stringerTest{
		input: &ExpectedConfigValue{arg0: "local", arg1: "httpd", arg2: "foo"},
		expected: `call to ConfigValue() which:
	- has node: local
	- has section: httpd
	- has key: foo`,
	})
	tests.Add("results", stringerTest{
		input: &ExpectedConfigValue{ret0: "baz"},
		expected: `call to ConfigValue() which:
	- has any node
	- has any section
	- has any key
	- should return: baz`,
	})

	tests.Run(t, testStringer)
}

func TestSetConfigValueString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedSetConfigValue{},
		expected: `call to SetConfigValue() which:
	- has any node
	- has any section
	- has any key
	- has any value`,
	})
	tests.Add("full", stringerTest{
		input: &ExpectedSetConfigValue{arg0: "local", arg1: "httpd", arg2: "foo", arg3: "bar"},
		expected: `call to SetConfigValue() which:
	- has node: local
	- has section: httpd
	- has key: foo
	- has value: bar`,
	})
	tests.Add("results", stringerTest{
		input: &ExpectedSetConfigValue{ret0: "baz"},
		expected: `call to SetConfigValue() which:
	- has any node
	- has any section
	- has any key
	- has any value
	- should return: baz`,
	})

	tests.Run(t, testStringer)
}

func TestDeleteConfigKeyString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedDeleteConfigKey{},
		expected: `call to DeleteConfigKey() which:
	- has any node
	- has any section
	- has any key`,
	})
	tests.Add("full", stringerTest{
		input: &ExpectedDeleteConfigKey{arg0: "local", arg1: "httpd", arg2: "foo"},
		expected: `call to DeleteConfigKey() which:
	- has node: local
	- has section: httpd
	- has key: foo`,
	})
	tests.Add("results", stringerTest{
		input: &ExpectedDeleteConfigKey{ret0: "baz"},
		expected: `call to DeleteConfigKey() which:
	- has any node
	- has any section
	- has any key
	- should return: baz`,
	})

	tests.Run(t, testStringer)
}

func TestRevsDiffString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedRevsDiff{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).RevsDiff() which:
	- has any revMap`,
	})
	tests.Add("revMap", stringerTest{
		input: &ExpectedRevsDiff{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: map[string][]string{"foo": {"1", "2"}}},
		expected: `call to DB(foo#0).RevsDiff() which:
	- with revMap: map[foo:[1 2]]`,
	})
	tests.Add("results", stringerTest{
		input: &ExpectedRevsDiff{
			commonExpectation: commonExpectation{db: &DB{name: "foo"}},
			ret0:              &Rows{},
		},
		expected: `call to DB(foo#0).RevsDiff() which:
	- has any revMap
	- should return: 0 results`,
	})

	tests.Run(t, testStringer)
}

func TestPartitionStatsString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input: &ExpectedPartitionStats{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).PartitionStats() which:
	- has any name`,
	})
	tests.Add("name", stringerTest{
		input: &ExpectedPartitionStats{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, arg0: "foo"},
		expected: `call to DB(foo#0).PartitionStats() which:
	- with name: foo`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedPartitionStats{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo err")}},
		expected: `call to DB(foo#0).PartitionStats() which:
	- has any name
	- should return error: foo err`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedPartitionStats{commonExpectation: commonExpectation{db: &DB{name: "foo"}, delay: time.Second}},
		expected: `call to DB(foo#0).PartitionStats() which:
	- has any name
	- should delay for: 1s`,
	})
	tests.Add("return", stringerTest{
		input: &ExpectedPartitionStats{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: &driver.PartitionStats{DBName: "foo"}},
		expected: `call to DB(foo#0).PartitionStats() which:
	- has any name
	- should return: {"DBName":"foo","DocCount":0,"DeletedDocCount":0,"Partition":"","ActiveSize":0,"ExternalSize":0,"RawResponse":null}`,
	})
	tests.Run(t, testStringer)
}

func TestSecurityString(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("empty", stringerTest{
		input:    &ExpectedSecurity{commonExpectation: commonExpectation{db: &DB{name: "foo"}}},
		expected: `call to DB(foo#0).Security()`,
	})
	tests.Add("error", stringerTest{
		input: &ExpectedSecurity{commonExpectation: commonExpectation{db: &DB{name: "foo"}, err: errors.New("foo err")}},
		expected: `call to DB(foo#0).Security() which:
	- should return error: foo err`,
	})
	tests.Add("delay", stringerTest{
		input: &ExpectedSecurity{commonExpectation: commonExpectation{db: &DB{name: "foo"}, delay: time.Second}},
		expected: `call to DB(foo#0).Security() which:
	- should delay for: 1s`,
	})
	tests.Add("return", stringerTest{
		input: &ExpectedSecurity{commonExpectation: commonExpectation{db: &DB{name: "foo"}}, ret0: &driver.Security{Admins: driver.Members{Names: []string{"bob", "alice"}}}},
		expected: `call to DB(foo#0).Security() which:
	- should return: {"admins":{"names":["bob","alice"]}}`,
	})
	tests.Run(t, testStringer)
}
