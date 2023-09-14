// Package mockdb provides a full Kivik driver implementation, for mocking in
// tests.
package mockdb

//go:generate go run ./gen
//go:generate gofmt -s -w clientexpectations_gen.go client_gen.go dbexpectations_gen.go db_gen.go clientmock_gen.go dbmock_gen.go
