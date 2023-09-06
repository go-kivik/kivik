// Package kivikmock provides a full Kivik driver implementation, for mocking
// unit tests.
package kivikmock

//go:generate go run ./gen ./gen/templates
//go:generate gofmt -s -w clientexpectations_gen.go client_gen.go dbexpectations_gen.go db_gen.go clientmock_gen.go dbmock_gen.go
