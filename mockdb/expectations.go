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

package mockdb

import (
	"context"
	"fmt"
	"time"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

func (e *ExpectedClose) String() string {
	extra := delayString(e.delay) + errorString(e.err)
	msg := "call to Close()"
	if extra != "" {
		msg += " which:" + extra
	}
	return msg
}

// String satisfies the fmt.Stringer interface.
func (e *ExpectedAllDBs) String() string {
	return "call to AllDBs() which:" +
		optionsString(e.options) +
		delayString(e.delay) +
		errorString(e.err)
}

func (e *ExpectedClusterSetup) String() string {
	msg := "call to ClusterSetup() which:"
	if e.arg0 == nil {
		msg += "\n\t- has any action"
	} else {
		msg += fmt.Sprintf("\n\t- has the action: %v", e.arg0)
	}
	msg += delayString(e.delay)
	msg += errorString(e.err)
	return msg
}

// String satisfies the fmt.Stringer interface
func (e *ExpectedClusterStatus) String() string {
	return "call to ClusterStatus() which:" +
		optionsString(e.options) +
		delayString(e.delay) +
		errorString(e.err)
}

func (e *ExpectedMembership) String() string {
	extra := delayString(e.delay) + errorString(e.err)
	msg := "call to Membership()"
	if extra != "" {
		msg += " which:" + extra
	}
	return msg
}

// WithAction specifies the action to be matched. Note that this expectation
// is compared with the actual action's marshaled JSON output, so it is not
// essential that the data types match exactly, in a Go sense.
func (e *ExpectedClusterSetup) WithAction(action interface{}) *ExpectedClusterSetup {
	e.arg0 = action
	return e
}

func (e *ExpectedDBExists) String() string {
	msg := "call to DBExists() which:" +
		fieldString("name", e.arg0) +
		optionsString(e.options) +
		delayString(e.delay)
	if e.err == nil {
		msg += fmt.Sprintf("\n\t- should return: %t", e.ret0)
	} else {
		msg += fmt.Sprintf("\n\t- should return error: %s", e.err)
	}
	return msg
}

// WithName sets the expectation that DBExists will be called with the provided
// name.
func (e *ExpectedDBExists) WithName(name string) *ExpectedDBExists {
	e.arg0 = name
	return e
}

func (e *ExpectedDestroyDB) String() string {
	return "call to DestroyDB() which:" +
		fieldString("name", e.arg0) +
		optionsString(e.options) +
		delayString(e.delay) +
		errorString(e.err)
}

// WithName sets the expectation that DestroyDB will be called with this name.
func (e *ExpectedDestroyDB) WithName(name string) *ExpectedDestroyDB {
	e.arg0 = name
	return e
}

func (e *ExpectedDBsStats) String() string {
	msg := "call to DBsStats() which:"
	if e.arg0 == nil {
		msg += "\n\t- has any names"
	} else {
		msg += fmt.Sprintf("\n\t- has names: %s", e.arg0)
	}
	return msg + delayString(e.delay) + errorString(e.err)
}

// WithNames sets the expectation that DBsStats will be called with these names.
func (e *ExpectedDBsStats) WithNames(names []string) *ExpectedDBsStats {
	e.arg0 = names
	return e
}

func (e *ExpectedAllDBsStats) String() string {
	return "call to AllDBsStats() which:" +
		optionsString(e.options) +
		delayString(e.delay) +
		errorString(e.err)
}

func (e *ExpectedPing) String() string {
	msg := "call to Ping()"
	extra := delayString(e.delay) + errorString(e.err)
	if extra != "" {
		msg += " which:" + extra
	}
	return msg
}

func (e *ExpectedSession) String() string {
	msg := "call to Session()"
	extra := ""
	if e.ret0 != nil {
		extra += fmt.Sprintf("\n\t- should return: %v", e.ret0)
	}
	extra += delayString(e.delay) + errorString(e.err)
	if extra != "" {
		msg += " which:" + extra
	}
	return msg
}

func (e *ExpectedVersion) String() string {
	msg := "call to Version()"
	extra := ""
	if e.ret0 != nil {
		extra += fmt.Sprintf("\n\t- should return: %v", e.ret0)
	}
	extra += delayString(e.delay) + errorString(e.err)
	if extra != "" {
		msg += " which:" + extra
	}
	return msg
}

func (e *ExpectedDB) String() string {
	msg := "call to DB() which:" +
		fieldString("name", e.arg0) +
		optionsString(e.options)
	if e.db != nil {
		msg += fmt.Sprintf("\n\t- should return database with %d expectations", e.db.expectations())
	}
	msg += delayString(e.delay)
	return msg
}

// WithName sets the expectation that DB() will be called with this name.
func (e *ExpectedDB) WithName(name string) *ExpectedDB {
	e.arg0 = name
	return e
}

// ExpectedCreateDB represents an expectation to call the CreateDB() method.
//
// Implementation note: Because kivik always calls DB() after a
// successful CreateDB() is executed, ExpectCreateDB() creates two
// expectations under the covers, one for the backend CreateDB() call,
// and one for the DB() call. If WillReturnError() is called, the DB() call
// expectation is removed.
type ExpectedCreateDB struct {
	commonExpectation
	arg0     string
	callback func(ctx context.Context, arg0 string, options driver.Options) error
}

func (e *ExpectedCreateDB) String() string {
	msg := "call to CreateDB() which:" +
		fieldString("name", e.arg0) +
		optionsString(e.options)
	if e.db != nil {
		msg += fmt.Sprintf("\n\t- should return database with %d expectations", e.db.expectations())
	}
	msg += delayString(e.delay)
	msg += errorString(e.err)
	return msg
}

func (e *ExpectedCreateDB) method(v bool) string {
	if !v {
		return "CreateDB()"
	}
	var name, options string
	if e.arg0 == "" {
		name = "?"
	} else {
		name = fmt.Sprintf("%q", e.arg0)
	}
	if e.options != nil {
		options = fmt.Sprintf(", %v", e.options)
	}
	return fmt.Sprintf("CreateDB(ctx, %s%s)", name, options)
}

func (e *ExpectedCreateDB) met(ex expectation) bool {
	exp := ex.(*ExpectedCreateDB)
	return exp.arg0 == "" || exp.arg0 == e.arg0
}

// WithOptions set the expectation that DB() will be called with these options.
func (e *ExpectedCreateDB) WithOptions(options ...kivik.Option) *ExpectedCreateDB {
	e.options = multiOptions{e.options, multiOptions(options)}
	return e
}

// WithName sets the expectation that DB() will be called with this name.
func (e *ExpectedCreateDB) WithName(name string) *ExpectedCreateDB {
	e.arg0 = name
	return e
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedCreateDB) WillExecute(cb func(ctx context.Context, arg0 string, options driver.Options) error) *ExpectedCreateDB {
	e.callback = cb
	return e
}

// WillReturnError sets the return value for the DB() call.
func (e *ExpectedCreateDB) WillReturnError(err error) *ExpectedCreateDB {
	e.err = err
	return e
}

// WillDelay will cause execution of DB() to delay by duration d.
func (e *ExpectedCreateDB) WillDelay(delay time.Duration) *ExpectedCreateDB {
	e.delay = delay
	return e
}

func (e *ExpectedDBUpdates) String() string {
	msg := "call to DBUpdates()"
	var extra string
	if e.ret0 != nil {
		extra += fmt.Sprintf("\n\t- should return: %d results", e.ret0.count())
	}
	extra += delayString(e.delay)
	extra += errorString(e.err)
	if extra != "" {
		msg += " which:" + extra
	}
	return msg
}

func (e *ExpectedConfig) String() string {
	msg := "call to Config() which:"
	msg += fieldString("node", e.arg0)
	if e.ret0 != nil {
		msg += fmt.Sprintf("\n\t- should return: %v", e.ret0)
	}
	msg += delayString(e.delay)
	msg += errorString(e.err)
	return msg
}

// WithNode sets the expected node.
func (e *ExpectedConfig) WithNode(node string) *ExpectedConfig {
	e.arg0 = node
	return e
}

func (e *ExpectedConfigSection) String() string {
	msg := "call to ConfigSection() which:" +
		fieldString("node", e.arg0) +
		fieldString("section", e.arg1)
	if e.ret0 != nil {
		msg += fmt.Sprintf("\n\t- should return: %v", e.ret0)
	}
	msg += delayString(e.delay)
	msg += errorString(e.err)
	return msg
}

// WithNode sets the expected node.
func (e *ExpectedConfigSection) WithNode(node string) *ExpectedConfigSection {
	e.arg0 = node
	return e
}

// WithSection sets the expected section.
func (e *ExpectedConfigSection) WithSection(section string) *ExpectedConfigSection {
	e.arg1 = section
	return e
}

func (e *ExpectedConfigValue) String() string {
	msg := "call to ConfigValue() which:" +
		fieldString("node", e.arg0) +
		fieldString("section", e.arg1) +
		fieldString("key", e.arg2)
	if e.ret0 != "" {
		msg += fmt.Sprintf("\n\t- should return: %v", e.ret0)
	}
	msg += delayString(e.delay)
	msg += errorString(e.err)
	return msg
}

// WithNode sets the expected node.
func (e *ExpectedConfigValue) WithNode(node string) *ExpectedConfigValue {
	e.arg0 = node
	return e
}

// WithSection sets the expected section.
func (e *ExpectedConfigValue) WithSection(section string) *ExpectedConfigValue {
	e.arg1 = section
	return e
}

// WithKey sets the expected key.
func (e *ExpectedConfigValue) WithKey(key string) *ExpectedConfigValue {
	e.arg2 = key
	return e
}

func (e *ExpectedSetConfigValue) String() string {
	msg := "call to SetConfigValue() which:" +
		fieldString("node", e.arg0) +
		fieldString("section", e.arg1) +
		fieldString("key", e.arg2) +
		fieldString("value", e.arg3)
	if e.ret0 != "" {
		msg += fmt.Sprintf("\n\t- should return: %v", e.ret0)
	}
	msg += delayString(e.delay)
	msg += errorString(e.err)
	return msg
}

// WithNode sets the expected node.
func (e *ExpectedSetConfigValue) WithNode(node string) *ExpectedSetConfigValue {
	e.arg0 = node
	return e
}

// WithSection sets the expected section.
func (e *ExpectedSetConfigValue) WithSection(section string) *ExpectedSetConfigValue {
	e.arg1 = section
	return e
}

// WithKey sets the expected key.
func (e *ExpectedSetConfigValue) WithKey(key string) *ExpectedSetConfigValue {
	e.arg2 = key
	return e
}

// WithValue sets the expected value.
func (e *ExpectedSetConfigValue) WithValue(value string) *ExpectedSetConfigValue {
	e.arg3 = value
	return e
}

func (e *ExpectedDeleteConfigKey) String() string {
	msg := "call to DeleteConfigKey() which:" +
		fieldString("node", e.arg0) +
		fieldString("section", e.arg1) +
		fieldString("key", e.arg2)
	if e.ret0 != "" {
		msg += fmt.Sprintf("\n\t- should return: %v", e.ret0)
	}
	msg += delayString(e.delay)
	msg += errorString(e.err)
	return msg
}

// WithNode sets the expected node.
func (e *ExpectedDeleteConfigKey) WithNode(node string) *ExpectedDeleteConfigKey {
	e.arg0 = node
	return e
}

// WithSection sets the expected section.
func (e *ExpectedDeleteConfigKey) WithSection(section string) *ExpectedDeleteConfigKey {
	e.arg1 = section
	return e
}

// WithKey sets the expected key.
func (e *ExpectedDeleteConfigKey) WithKey(key string) *ExpectedDeleteConfigKey {
	e.arg2 = key
	return e
}

func (e *ExpectedReplicate) String() string {
	msg := "call to Replicate() which:" +
		fieldString("target", e.arg0) +
		fieldString("source", e.arg1) +
		optionsString(e.options)
	if e.ret0 != nil {
		msg += fmt.Sprintf("\n\t- should return: %v", jsonDoc(e.ret0))
	}
	return msg +
		delayString(e.delay) +
		errorString(e.err)
}

// WithSource sets the expected source.
func (e *ExpectedReplicate) WithSource(source string) *ExpectedReplicate {
	e.arg1 = source
	return e
}

// WithTarget sets the expected target.
func (e *ExpectedReplicate) WithTarget(target string) *ExpectedReplicate {
	e.arg0 = target
	return e
}

func (e *ExpectedGetReplications) String() string {
	msg := "call to GetReplications() which:" +
		optionsString(e.options)
	if l := len(e.ret0); l > 0 {
		msg += fmt.Sprintf("\n\t- should return: %d results", l)
	}
	return msg +
		delayString(e.delay) +
		errorString(e.err)
}
