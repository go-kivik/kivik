[![Build Status](https://travis-ci.org/go-kivik/kivikmock.svg?branch=master)](https://travis-ci.org/go-kivik/kivikmock) [![Codecov](https://img.shields.io/codecov/c/github/go-kivik/kivikmock.svg?style=flat)](https://codecov.io/gh/go-kivik/kivikmock) [![Go Report Card](https://goreportcard.com/badge/github.com/go-kivik/kivikmock)](https://goreportcard.com/report/github.com/go-kivik/kivikmock) [![GoDoc](https://godoc.org/github.com/go-kivik/kivikmock?status.svg)](http://godoc.org/github.com/go-kivik/kivikmock) [![Website](https://img.shields.io/website-up-down-green-red/http/shields.io.svg?label=website&colorB=007fff)](http://kivik.io)

# Kivikmock

Package **kivikmock** is a mock library implementing a Kivik driver.

This package is heavily influenced by [github.com/DATA-DOG/go-sqlmock](https://github.com/DATA-DOG/go-sqlmock), the SQL mock driver from [Datadog](https://www.datadoghq.com/).

# Usage

To use this package, in your `*_test.go` file, create a mock Kivik connection:

    client, mock, err := kivikmock.New()
    if err != nil {
        panic(err)
    }

The returned `client` object is a `*kivik.Client`, and can be passed to your
methods to be tested.   `mock` is used to control the execution of the mock
driver, by setting expectations.  To test a function which fetches a user,
for example, you might do something like this:

    func TestGetUser(t *testing.T) {
        client, mock, err := kivikmock.New()
        if err != nil {
            t.Fatal(err)
        }

        mock.ExpectDB().WithName("_users").WillReturn(mock.NewDB().
            ExpectGet().WithDocID("bob").
                WillReturn(kivikmock.DocumentT(t, `{"_id":"org.couchdb.user:bob"}`)),
        )
        user, err := GetUser(client, "bob")
        if err != nil {
            t.Error(err)
        }
        // other validation
    }

# Versions

This package targets the unstable release of Kivik.

## License

This software is released under the terms of the Apache 2.0 license. See
LICENCE.md, or read the [full license](http://www.apache.org/licenses/LICENSE-2.0).
