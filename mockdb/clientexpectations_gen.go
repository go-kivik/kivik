/* This file is auto-generated. Do not edit it! */

package mockdb

import (
	"context"
	"fmt"
	"reflect"
	"time"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

var _ = &driver.Attachment{}
var _ = reflect.Int

// ExpectedClose represents an expectation for a call to Close().
type ExpectedClose struct {
	commonExpectation
	callback func() error
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedClose) WillExecute(cb func() error) *ExpectedClose {
	e.callback = cb
	return e
}

// WillReturnError sets the error value that will be returned by the call to Close().
func (e *ExpectedClose) WillReturnError(err error) *ExpectedClose {
	e.err = err
	return e
}

func (e *ExpectedClose) met(_ expectation) bool {
	return true
}

func (e *ExpectedClose) method(v bool) string {
	if !v {
		return "Close()"
	}
	return fmt.Sprintf("Close()")
}

// ExpectedClusterSetup represents an expectation for a call to ClusterSetup().
type ExpectedClusterSetup struct {
	commonExpectation
	callback func(ctx context.Context, arg0 interface{}) error
	arg0     interface{}
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedClusterSetup) WillExecute(cb func(ctx context.Context, arg0 interface{}) error) *ExpectedClusterSetup {
	e.callback = cb
	return e
}

// WillReturnError sets the error value that will be returned by the call to ClusterSetup().
func (e *ExpectedClusterSetup) WillReturnError(err error) *ExpectedClusterSetup {
	e.err = err
	return e
}

// WillDelay causes the call to ClusterSetup() to delay.
func (e *ExpectedClusterSetup) WillDelay(delay time.Duration) *ExpectedClusterSetup {
	e.delay = delay
	return e
}

func (e *ExpectedClusterSetup) met(ex expectation) bool {
	exp := ex.(*ExpectedClusterSetup)
	if exp.arg0 != nil && !jsonMeets(exp.arg0, e.arg0) {
		return false
	}
	return true
}

func (e *ExpectedClusterSetup) method(v bool) string {
	if !v {
		return "ClusterSetup()"
	}
	arg0 := "?"
	if e.arg0 != nil {
		arg0 = fmt.Sprintf("%v", e.arg0)
	}
	return fmt.Sprintf("ClusterSetup(ctx, %s)", arg0)
}

// ExpectedConfigValue represents an expectation for a call to ConfigValue().
type ExpectedConfigValue struct {
	commonExpectation
	callback func(ctx context.Context, arg0 string, arg1 string, arg2 string) (string, error)
	arg0     string
	arg1     string
	arg2     string
	ret0     string
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedConfigValue) WillExecute(cb func(ctx context.Context, arg0 string, arg1 string, arg2 string) (string, error)) *ExpectedConfigValue {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to ConfigValue().
func (e *ExpectedConfigValue) WillReturn(ret0 string) *ExpectedConfigValue {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to ConfigValue().
func (e *ExpectedConfigValue) WillReturnError(err error) *ExpectedConfigValue {
	e.err = err
	return e
}

// WillDelay causes the call to ConfigValue() to delay.
func (e *ExpectedConfigValue) WillDelay(delay time.Duration) *ExpectedConfigValue {
	e.delay = delay
	return e
}

func (e *ExpectedConfigValue) met(ex expectation) bool {
	exp := ex.(*ExpectedConfigValue)
	if exp.arg0 != "" && exp.arg0 != e.arg0 {
		return false
	}
	if exp.arg1 != "" && exp.arg1 != e.arg1 {
		return false
	}
	if exp.arg2 != "" && exp.arg2 != e.arg2 {
		return false
	}
	return true
}

func (e *ExpectedConfigValue) method(v bool) string {
	if !v {
		return "ConfigValue()"
	}
	arg0, arg1, arg2 := "?", "?", "?"
	if e.arg0 != "" {
		arg0 = fmt.Sprintf("%q", e.arg0)
	}
	if e.arg1 != "" {
		arg1 = fmt.Sprintf("%q", e.arg1)
	}
	if e.arg2 != "" {
		arg2 = fmt.Sprintf("%q", e.arg2)
	}
	return fmt.Sprintf("ConfigValue(ctx, %s, %s, %s)", arg0, arg1, arg2)
}

// ExpectedDeleteConfigKey represents an expectation for a call to DeleteConfigKey().
type ExpectedDeleteConfigKey struct {
	commonExpectation
	callback func(ctx context.Context, arg0 string, arg1 string, arg2 string) (string, error)
	arg0     string
	arg1     string
	arg2     string
	ret0     string
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedDeleteConfigKey) WillExecute(cb func(ctx context.Context, arg0 string, arg1 string, arg2 string) (string, error)) *ExpectedDeleteConfigKey {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to DeleteConfigKey().
func (e *ExpectedDeleteConfigKey) WillReturn(ret0 string) *ExpectedDeleteConfigKey {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to DeleteConfigKey().
func (e *ExpectedDeleteConfigKey) WillReturnError(err error) *ExpectedDeleteConfigKey {
	e.err = err
	return e
}

// WillDelay causes the call to DeleteConfigKey() to delay.
func (e *ExpectedDeleteConfigKey) WillDelay(delay time.Duration) *ExpectedDeleteConfigKey {
	e.delay = delay
	return e
}

func (e *ExpectedDeleteConfigKey) met(ex expectation) bool {
	exp := ex.(*ExpectedDeleteConfigKey)
	if exp.arg0 != "" && exp.arg0 != e.arg0 {
		return false
	}
	if exp.arg1 != "" && exp.arg1 != e.arg1 {
		return false
	}
	if exp.arg2 != "" && exp.arg2 != e.arg2 {
		return false
	}
	return true
}

func (e *ExpectedDeleteConfigKey) method(v bool) string {
	if !v {
		return "DeleteConfigKey()"
	}
	arg0, arg1, arg2 := "?", "?", "?"
	if e.arg0 != "" {
		arg0 = fmt.Sprintf("%q", e.arg0)
	}
	if e.arg1 != "" {
		arg1 = fmt.Sprintf("%q", e.arg1)
	}
	if e.arg2 != "" {
		arg2 = fmt.Sprintf("%q", e.arg2)
	}
	return fmt.Sprintf("DeleteConfigKey(ctx, %s, %s, %s)", arg0, arg1, arg2)
}

// ExpectedPing represents an expectation for a call to Ping().
type ExpectedPing struct {
	commonExpectation
	callback func(ctx context.Context) (bool, error)
	ret0     bool
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedPing) WillExecute(cb func(ctx context.Context) (bool, error)) *ExpectedPing {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to Ping().
func (e *ExpectedPing) WillReturn(ret0 bool) *ExpectedPing {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to Ping().
func (e *ExpectedPing) WillReturnError(err error) *ExpectedPing {
	e.err = err
	return e
}

// WillDelay causes the call to Ping() to delay.
func (e *ExpectedPing) WillDelay(delay time.Duration) *ExpectedPing {
	e.delay = delay
	return e
}

func (e *ExpectedPing) met(_ expectation) bool {
	return true
}

func (e *ExpectedPing) method(v bool) string {
	if !v {
		return "Ping()"
	}
	return fmt.Sprintf("Ping(ctx)")
}

// ExpectedSetConfigValue represents an expectation for a call to SetConfigValue().
type ExpectedSetConfigValue struct {
	commonExpectation
	callback func(ctx context.Context, arg0 string, arg1 string, arg2 string, arg3 string) (string, error)
	arg0     string
	arg1     string
	arg2     string
	arg3     string
	ret0     string
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedSetConfigValue) WillExecute(cb func(ctx context.Context, arg0 string, arg1 string, arg2 string, arg3 string) (string, error)) *ExpectedSetConfigValue {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to SetConfigValue().
func (e *ExpectedSetConfigValue) WillReturn(ret0 string) *ExpectedSetConfigValue {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to SetConfigValue().
func (e *ExpectedSetConfigValue) WillReturnError(err error) *ExpectedSetConfigValue {
	e.err = err
	return e
}

// WillDelay causes the call to SetConfigValue() to delay.
func (e *ExpectedSetConfigValue) WillDelay(delay time.Duration) *ExpectedSetConfigValue {
	e.delay = delay
	return e
}

func (e *ExpectedSetConfigValue) met(ex expectation) bool {
	exp := ex.(*ExpectedSetConfigValue)
	if exp.arg0 != "" && exp.arg0 != e.arg0 {
		return false
	}
	if exp.arg1 != "" && exp.arg1 != e.arg1 {
		return false
	}
	if exp.arg2 != "" && exp.arg2 != e.arg2 {
		return false
	}
	if exp.arg3 != "" && exp.arg3 != e.arg3 {
		return false
	}
	return true
}

func (e *ExpectedSetConfigValue) method(v bool) string {
	if !v {
		return "SetConfigValue()"
	}
	arg0, arg1, arg2, arg3 := "?", "?", "?", "?"
	if e.arg0 != "" {
		arg0 = fmt.Sprintf("%q", e.arg0)
	}
	if e.arg1 != "" {
		arg1 = fmt.Sprintf("%q", e.arg1)
	}
	if e.arg2 != "" {
		arg2 = fmt.Sprintf("%q", e.arg2)
	}
	if e.arg3 != "" {
		arg3 = fmt.Sprintf("%q", e.arg3)
	}
	return fmt.Sprintf("SetConfigValue(ctx, %s, %s, %s, %s)", arg0, arg1, arg2, arg3)
}

// ExpectedAllDBs represents an expectation for a call to AllDBs().
type ExpectedAllDBs struct {
	commonExpectation
	callback func(ctx context.Context, options driver.Options) ([]string, error)
	ret0     []string
}

// WithOptions sets the expected options for the call to AllDBs().
func (e *ExpectedAllDBs) WithOptions(options driver.Options) *ExpectedAllDBs {
	e.options = toLegacyOptions(options)
	return e
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedAllDBs) WillExecute(cb func(ctx context.Context, options driver.Options) ([]string, error)) *ExpectedAllDBs {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to AllDBs().
func (e *ExpectedAllDBs) WillReturn(ret0 []string) *ExpectedAllDBs {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to AllDBs().
func (e *ExpectedAllDBs) WillReturnError(err error) *ExpectedAllDBs {
	e.err = err
	return e
}

// WillDelay causes the call to AllDBs() to delay.
func (e *ExpectedAllDBs) WillDelay(delay time.Duration) *ExpectedAllDBs {
	e.delay = delay
	return e
}

func (e *ExpectedAllDBs) met(_ expectation) bool {
	return true
}

func (e *ExpectedAllDBs) method(v bool) string {
	if !v {
		return "AllDBs()"
	}
	options := defaultOptionPlaceholder
	if e.options != nil {
		options = fmt.Sprintf("%v", e.options)
	}
	return fmt.Sprintf("AllDBs(ctx, %s)", options)
}

// ExpectedClusterStatus represents an expectation for a call to ClusterStatus().
type ExpectedClusterStatus struct {
	commonExpectation
	callback func(ctx context.Context, options map[string]interface{}) (string, error)
	ret0     string
}

// WithOptions sets the expected options for the call to ClusterStatus().
func (e *ExpectedClusterStatus) WithOptions(options kivik.Options) *ExpectedClusterStatus {
	e.options = options
	return e
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedClusterStatus) WillExecute(cb func(ctx context.Context, options map[string]interface{}) (string, error)) *ExpectedClusterStatus {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to ClusterStatus().
func (e *ExpectedClusterStatus) WillReturn(ret0 string) *ExpectedClusterStatus {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to ClusterStatus().
func (e *ExpectedClusterStatus) WillReturnError(err error) *ExpectedClusterStatus {
	e.err = err
	return e
}

// WillDelay causes the call to ClusterStatus() to delay.
func (e *ExpectedClusterStatus) WillDelay(delay time.Duration) *ExpectedClusterStatus {
	e.delay = delay
	return e
}

func (e *ExpectedClusterStatus) met(_ expectation) bool {
	return true
}

func (e *ExpectedClusterStatus) method(v bool) string {
	if !v {
		return "ClusterStatus()"
	}
	options := defaultOptionPlaceholder
	if e.options != nil {
		options = fmt.Sprintf("%v", e.options)
	}
	return fmt.Sprintf("ClusterStatus(ctx, %s)", options)
}

// ExpectedConfig represents an expectation for a call to Config().
type ExpectedConfig struct {
	commonExpectation
	callback func(ctx context.Context, arg0 string) (driver.Config, error)
	arg0     string
	ret0     driver.Config
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedConfig) WillExecute(cb func(ctx context.Context, arg0 string) (driver.Config, error)) *ExpectedConfig {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to Config().
func (e *ExpectedConfig) WillReturn(ret0 driver.Config) *ExpectedConfig {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to Config().
func (e *ExpectedConfig) WillReturnError(err error) *ExpectedConfig {
	e.err = err
	return e
}

// WillDelay causes the call to Config() to delay.
func (e *ExpectedConfig) WillDelay(delay time.Duration) *ExpectedConfig {
	e.delay = delay
	return e
}

func (e *ExpectedConfig) met(ex expectation) bool {
	exp := ex.(*ExpectedConfig)
	if exp.arg0 != "" && exp.arg0 != e.arg0 {
		return false
	}
	return true
}

func (e *ExpectedConfig) method(v bool) string {
	if !v {
		return "Config()"
	}
	arg0 := "?"
	if e.arg0 != "" {
		arg0 = fmt.Sprintf("%q", e.arg0)
	}
	return fmt.Sprintf("Config(ctx, %s)", arg0)
}

// ExpectedConfigSection represents an expectation for a call to ConfigSection().
type ExpectedConfigSection struct {
	commonExpectation
	callback func(ctx context.Context, arg0 string, arg1 string) (driver.ConfigSection, error)
	arg0     string
	arg1     string
	ret0     driver.ConfigSection
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedConfigSection) WillExecute(cb func(ctx context.Context, arg0 string, arg1 string) (driver.ConfigSection, error)) *ExpectedConfigSection {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to ConfigSection().
func (e *ExpectedConfigSection) WillReturn(ret0 driver.ConfigSection) *ExpectedConfigSection {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to ConfigSection().
func (e *ExpectedConfigSection) WillReturnError(err error) *ExpectedConfigSection {
	e.err = err
	return e
}

// WillDelay causes the call to ConfigSection() to delay.
func (e *ExpectedConfigSection) WillDelay(delay time.Duration) *ExpectedConfigSection {
	e.delay = delay
	return e
}

func (e *ExpectedConfigSection) met(ex expectation) bool {
	exp := ex.(*ExpectedConfigSection)
	if exp.arg0 != "" && exp.arg0 != e.arg0 {
		return false
	}
	if exp.arg1 != "" && exp.arg1 != e.arg1 {
		return false
	}
	return true
}

func (e *ExpectedConfigSection) method(v bool) string {
	if !v {
		return "ConfigSection()"
	}
	arg0, arg1 := "?", "?"
	if e.arg0 != "" {
		arg0 = fmt.Sprintf("%q", e.arg0)
	}
	if e.arg1 != "" {
		arg1 = fmt.Sprintf("%q", e.arg1)
	}
	return fmt.Sprintf("ConfigSection(ctx, %s, %s)", arg0, arg1)
}

// ExpectedDB represents an expectation for a call to DB().
type ExpectedDB struct {
	commonExpectation
	callback func(arg0 string, options map[string]interface{}) (driver.DB, error)
	arg0     string
	ret0     *DB
}

// WithOptions sets the expected options for the call to DB().
func (e *ExpectedDB) WithOptions(options kivik.Options) *ExpectedDB {
	e.options = options
	return e
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedDB) WillExecute(cb func(arg0 string, options map[string]interface{}) (driver.DB, error)) *ExpectedDB {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to DB().
func (e *ExpectedDB) WillReturn(ret0 *DB) *ExpectedDB {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to DB().
func (e *ExpectedDB) WillReturnError(err error) *ExpectedDB {
	e.err = err
	return e
}

func (e *ExpectedDB) met(ex expectation) bool {
	exp := ex.(*ExpectedDB)
	if exp.arg0 != "" && exp.arg0 != e.arg0 {
		return false
	}
	return true
}

func (e *ExpectedDB) method(v bool) string {
	if !v {
		return "DB()"
	}
	arg0, options := "?", defaultOptionPlaceholder
	if e.arg0 != "" {
		arg0 = fmt.Sprintf("%q", e.arg0)
	}
	if e.options != nil {
		options = fmt.Sprintf("%v", e.options)
	}
	return fmt.Sprintf("DB(%s, %s)", arg0, options)
}

// ExpectedDBExists represents an expectation for a call to DBExists().
type ExpectedDBExists struct {
	commonExpectation
	callback func(ctx context.Context, arg0 string, options map[string]interface{}) (bool, error)
	arg0     string
	ret0     bool
}

// WithOptions sets the expected options for the call to DBExists().
func (e *ExpectedDBExists) WithOptions(options kivik.Options) *ExpectedDBExists {
	e.options = options
	return e
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedDBExists) WillExecute(cb func(ctx context.Context, arg0 string, options map[string]interface{}) (bool, error)) *ExpectedDBExists {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to DBExists().
func (e *ExpectedDBExists) WillReturn(ret0 bool) *ExpectedDBExists {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to DBExists().
func (e *ExpectedDBExists) WillReturnError(err error) *ExpectedDBExists {
	e.err = err
	return e
}

// WillDelay causes the call to DBExists() to delay.
func (e *ExpectedDBExists) WillDelay(delay time.Duration) *ExpectedDBExists {
	e.delay = delay
	return e
}

func (e *ExpectedDBExists) met(ex expectation) bool {
	exp := ex.(*ExpectedDBExists)
	if exp.arg0 != "" && exp.arg0 != e.arg0 {
		return false
	}
	return true
}

func (e *ExpectedDBExists) method(v bool) string {
	if !v {
		return "DBExists()"
	}
	arg0, options := "?", defaultOptionPlaceholder
	if e.arg0 != "" {
		arg0 = fmt.Sprintf("%q", e.arg0)
	}
	if e.options != nil {
		options = fmt.Sprintf("%v", e.options)
	}
	return fmt.Sprintf("DBExists(ctx, %s, %s)", arg0, options)
}

// ExpectedDBUpdates represents an expectation for a call to DBUpdates().
type ExpectedDBUpdates struct {
	commonExpectation
	callback func(ctx context.Context, options map[string]interface{}) (driver.DBUpdates, error)
	ret0     *Updates
}

// WithOptions sets the expected options for the call to DBUpdates().
func (e *ExpectedDBUpdates) WithOptions(options kivik.Options) *ExpectedDBUpdates {
	e.options = options
	return e
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedDBUpdates) WillExecute(cb func(ctx context.Context, options map[string]interface{}) (driver.DBUpdates, error)) *ExpectedDBUpdates {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to DBUpdates().
func (e *ExpectedDBUpdates) WillReturn(ret0 *Updates) *ExpectedDBUpdates {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to DBUpdates().
func (e *ExpectedDBUpdates) WillReturnError(err error) *ExpectedDBUpdates {
	e.err = err
	return e
}

// WillDelay causes the call to DBUpdates() to delay.
func (e *ExpectedDBUpdates) WillDelay(delay time.Duration) *ExpectedDBUpdates {
	e.delay = delay
	return e
}

func (e *ExpectedDBUpdates) met(_ expectation) bool {
	return true
}

func (e *ExpectedDBUpdates) method(v bool) string {
	if !v {
		return "DBUpdates()"
	}
	options := defaultOptionPlaceholder
	if e.options != nil {
		options = fmt.Sprintf("%v", e.options)
	}
	return fmt.Sprintf("DBUpdates(ctx, %s)", options)
}

// ExpectedDBsStats represents an expectation for a call to DBsStats().
type ExpectedDBsStats struct {
	commonExpectation
	callback func(ctx context.Context, arg0 []string) ([]*driver.DBStats, error)
	arg0     []string
	ret0     []*driver.DBStats
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedDBsStats) WillExecute(cb func(ctx context.Context, arg0 []string) ([]*driver.DBStats, error)) *ExpectedDBsStats {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to DBsStats().
func (e *ExpectedDBsStats) WillReturn(ret0 []*driver.DBStats) *ExpectedDBsStats {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to DBsStats().
func (e *ExpectedDBsStats) WillReturnError(err error) *ExpectedDBsStats {
	e.err = err
	return e
}

// WillDelay causes the call to DBsStats() to delay.
func (e *ExpectedDBsStats) WillDelay(delay time.Duration) *ExpectedDBsStats {
	e.delay = delay
	return e
}

func (e *ExpectedDBsStats) met(ex expectation) bool {
	exp := ex.(*ExpectedDBsStats)
	if exp.arg0 != nil && !reflect.DeepEqual(exp.arg0, e.arg0) {
		return false
	}
	return true
}

func (e *ExpectedDBsStats) method(v bool) string {
	if !v {
		return "DBsStats()"
	}
	arg0 := "?"
	if e.arg0 != nil {
		arg0 = fmt.Sprintf("%v", e.arg0)
	}
	return fmt.Sprintf("DBsStats(ctx, %s)", arg0)
}

// ExpectedDestroyDB represents an expectation for a call to DestroyDB().
type ExpectedDestroyDB struct {
	commonExpectation
	callback func(ctx context.Context, arg0 string, options map[string]interface{}) error
	arg0     string
}

// WithOptions sets the expected options for the call to DestroyDB().
func (e *ExpectedDestroyDB) WithOptions(options kivik.Options) *ExpectedDestroyDB {
	e.options = options
	return e
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedDestroyDB) WillExecute(cb func(ctx context.Context, arg0 string, options map[string]interface{}) error) *ExpectedDestroyDB {
	e.callback = cb
	return e
}

// WillReturnError sets the error value that will be returned by the call to DestroyDB().
func (e *ExpectedDestroyDB) WillReturnError(err error) *ExpectedDestroyDB {
	e.err = err
	return e
}

// WillDelay causes the call to DestroyDB() to delay.
func (e *ExpectedDestroyDB) WillDelay(delay time.Duration) *ExpectedDestroyDB {
	e.delay = delay
	return e
}

func (e *ExpectedDestroyDB) met(ex expectation) bool {
	exp := ex.(*ExpectedDestroyDB)
	if exp.arg0 != "" && exp.arg0 != e.arg0 {
		return false
	}
	return true
}

func (e *ExpectedDestroyDB) method(v bool) string {
	if !v {
		return "DestroyDB()"
	}
	arg0, options := "?", defaultOptionPlaceholder
	if e.arg0 != "" {
		arg0 = fmt.Sprintf("%q", e.arg0)
	}
	if e.options != nil {
		options = fmt.Sprintf("%v", e.options)
	}
	return fmt.Sprintf("DestroyDB(ctx, %s, %s)", arg0, options)
}

// ExpectedGetReplications represents an expectation for a call to GetReplications().
type ExpectedGetReplications struct {
	commonExpectation
	callback func(ctx context.Context, options map[string]interface{}) ([]driver.Replication, error)
	ret0     []*Replication
}

// WithOptions sets the expected options for the call to GetReplications().
func (e *ExpectedGetReplications) WithOptions(options kivik.Options) *ExpectedGetReplications {
	e.options = options
	return e
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedGetReplications) WillExecute(cb func(ctx context.Context, options map[string]interface{}) ([]driver.Replication, error)) *ExpectedGetReplications {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to GetReplications().
func (e *ExpectedGetReplications) WillReturn(ret0 []*Replication) *ExpectedGetReplications {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to GetReplications().
func (e *ExpectedGetReplications) WillReturnError(err error) *ExpectedGetReplications {
	e.err = err
	return e
}

// WillDelay causes the call to GetReplications() to delay.
func (e *ExpectedGetReplications) WillDelay(delay time.Duration) *ExpectedGetReplications {
	e.delay = delay
	return e
}

func (e *ExpectedGetReplications) met(_ expectation) bool {
	return true
}

func (e *ExpectedGetReplications) method(v bool) string {
	if !v {
		return "GetReplications()"
	}
	options := defaultOptionPlaceholder
	if e.options != nil {
		options = fmt.Sprintf("%v", e.options)
	}
	return fmt.Sprintf("GetReplications(ctx, %s)", options)
}

// ExpectedMembership represents an expectation for a call to Membership().
type ExpectedMembership struct {
	commonExpectation
	callback func(ctx context.Context) (*driver.ClusterMembership, error)
	ret0     *driver.ClusterMembership
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedMembership) WillExecute(cb func(ctx context.Context) (*driver.ClusterMembership, error)) *ExpectedMembership {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to Membership().
func (e *ExpectedMembership) WillReturn(ret0 *driver.ClusterMembership) *ExpectedMembership {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to Membership().
func (e *ExpectedMembership) WillReturnError(err error) *ExpectedMembership {
	e.err = err
	return e
}

// WillDelay causes the call to Membership() to delay.
func (e *ExpectedMembership) WillDelay(delay time.Duration) *ExpectedMembership {
	e.delay = delay
	return e
}

func (e *ExpectedMembership) met(_ expectation) bool {
	return true
}

func (e *ExpectedMembership) method(v bool) string {
	if !v {
		return "Membership()"
	}
	return fmt.Sprintf("Membership(ctx)")
}

// ExpectedReplicate represents an expectation for a call to Replicate().
type ExpectedReplicate struct {
	commonExpectation
	callback func(ctx context.Context, arg0 string, arg1 string, options map[string]interface{}) (driver.Replication, error)
	arg0     string
	arg1     string
	ret0     *Replication
}

// WithOptions sets the expected options for the call to Replicate().
func (e *ExpectedReplicate) WithOptions(options kivik.Options) *ExpectedReplicate {
	e.options = options
	return e
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedReplicate) WillExecute(cb func(ctx context.Context, arg0 string, arg1 string, options map[string]interface{}) (driver.Replication, error)) *ExpectedReplicate {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to Replicate().
func (e *ExpectedReplicate) WillReturn(ret0 *Replication) *ExpectedReplicate {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to Replicate().
func (e *ExpectedReplicate) WillReturnError(err error) *ExpectedReplicate {
	e.err = err
	return e
}

// WillDelay causes the call to Replicate() to delay.
func (e *ExpectedReplicate) WillDelay(delay time.Duration) *ExpectedReplicate {
	e.delay = delay
	return e
}

func (e *ExpectedReplicate) met(ex expectation) bool {
	exp := ex.(*ExpectedReplicate)
	if exp.arg0 != "" && exp.arg0 != e.arg0 {
		return false
	}
	if exp.arg1 != "" && exp.arg1 != e.arg1 {
		return false
	}
	return true
}

func (e *ExpectedReplicate) method(v bool) string {
	if !v {
		return "Replicate()"
	}
	arg0, arg1, options := "?", "?", defaultOptionPlaceholder
	if e.arg0 != "" {
		arg0 = fmt.Sprintf("%q", e.arg0)
	}
	if e.arg1 != "" {
		arg1 = fmt.Sprintf("%q", e.arg1)
	}
	if e.options != nil {
		options = fmt.Sprintf("%v", e.options)
	}
	return fmt.Sprintf("Replicate(ctx, %s, %s, %s)", arg0, arg1, options)
}

// ExpectedSession represents an expectation for a call to Session().
type ExpectedSession struct {
	commonExpectation
	callback func(ctx context.Context) (*driver.Session, error)
	ret0     *driver.Session
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedSession) WillExecute(cb func(ctx context.Context) (*driver.Session, error)) *ExpectedSession {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to Session().
func (e *ExpectedSession) WillReturn(ret0 *driver.Session) *ExpectedSession {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to Session().
func (e *ExpectedSession) WillReturnError(err error) *ExpectedSession {
	e.err = err
	return e
}

// WillDelay causes the call to Session() to delay.
func (e *ExpectedSession) WillDelay(delay time.Duration) *ExpectedSession {
	e.delay = delay
	return e
}

func (e *ExpectedSession) met(_ expectation) bool {
	return true
}

func (e *ExpectedSession) method(v bool) string {
	if !v {
		return "Session()"
	}
	return fmt.Sprintf("Session(ctx)")
}

// ExpectedVersion represents an expectation for a call to Version().
type ExpectedVersion struct {
	commonExpectation
	callback func(ctx context.Context) (*driver.Version, error)
	ret0     *driver.Version
}

// WillExecute sets a callback function to be called with any inputs to the
// original function. Any values returned by the callback will be returned as
// if generated by the driver.
func (e *ExpectedVersion) WillExecute(cb func(ctx context.Context) (*driver.Version, error)) *ExpectedVersion {
	e.callback = cb
	return e
}

// WillReturn sets the values that will be returned by the call to Version().
func (e *ExpectedVersion) WillReturn(ret0 *driver.Version) *ExpectedVersion {
	e.ret0 = ret0
	return e
}

// WillReturnError sets the error value that will be returned by the call to Version().
func (e *ExpectedVersion) WillReturnError(err error) *ExpectedVersion {
	e.err = err
	return e
}

// WillDelay causes the call to Version() to delay.
func (e *ExpectedVersion) WillDelay(delay time.Duration) *ExpectedVersion {
	e.delay = delay
	return e
}

func (e *ExpectedVersion) met(_ expectation) bool {
	return true
}

func (e *ExpectedVersion) method(v bool) string {
	if !v {
		return "Version()"
	}
	return fmt.Sprintf("Version(ctx)")
}
