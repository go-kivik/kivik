# kivik CLI tool

The `kivik` CLI tool allows interacting with CouchDB manually or via shell scripts, without the toil of building complex queries with [curl](https://curl.se/).

## Motivation

While debugging, experimenting with, or administrating CouchDB, I found the repetitive use of long `curl` commands to be toilsome. I've also found the flexibliyt of tools like `kubectl` to be informative (i.e. supporting both JSON, YAML, and arbitrary output formats). And finally, I have often longed for a simple CouchDB analog to tools like `pg_dump` and `pg_restore`.

This tool aims to satisfy all of these itches.

## Status

This is still a work in progress. Some features are still completely unimplemented (see the [TODO list](TODO.md)). There are also likely bugs, as well as inconsistencies. [Bug reports](https://github.com/go-kivik/xkivik/issues) are very welcome!

## Installation

`kivik` is written in Go, and as such can be installed on any system with a working Go 1.14 or newer compiler:

```shell
go get github.com/go-kivik/xkivik/cmd/kivik
```

You may also download precompiled binaries for popular architectures from the [releases page](https://github.com/go-kivik/xkivik/releases).

## Basic usage

In general, `kivik` aims to provide an interface consistent with CouchDB. Most commands can be accessed in one of two ways:

```shell
kivik <verb> <command> <partial url> [arguments]
```

Or:

```shell
kivik <verb> <full url> [arguments]
```

This may be made more clear with an example.  The following two commands are equivalent:

```shell
kivik get document http://localhost:5984/db/SpaghettiWithMeatballs
kivik get http://localhost:5984/db/SpaghettiWithMeatballs
```

So why would you ever prefer the longer version? Because of the ability for `kivik` to understand a default context.

## Context

`kivik` understands the concept of contexts (the term and concept are borrowed from `kubectl`).  In the simplest form, it allows you to specify a default CouchDB server (and optionally database), so that they needn't be specified for every command.  This can be done in one of two ways:  Via the config file, or via environment variables.

(This feature still needs additional work, to support use of the non-default context.)

### config.yaml

The default config file location is `~/.kivik/config`, but you may specify any file with the `--config` flag.  The file is expected to be in YAML format, and is of the following format:

```yaml
contexts:
- context:
  name: short           # The name of the context
  scheme: https         # The schema (http, https, or file are supported)
  user: admin           # Username for authentication
  password: abc123      # Password for authentication
  host: localhost:5984  # Host and port of CouchDB server
  database: somepath/db # The path to a database
# A context may be specified as a single, compact DSN as in the below
# example as well.
- context:
  name: long
  dsn: https://admin:abc123@localhost:5984/somepath/db
default: short  # The name of the default context
```

If the config file contains only a single context, it is automatically used as the default.

You can also provide a context via the environment, by setting the `KIVIKDSN` environment variable.

```shell
KIVIKDSN=https://admin:abc123@localhost:5984/somepath/db
```

When a `kivik` command is executed, the DSN value provided on the command-line is merged with the default context to determine the actual DSN to use.  This means that if your default context contains a full path to a database, then you do not need to provide the database name on the `kivik` command line (unless you wish to interact with a different database).

Let's illustrate by expanding on our original example, this time with a context set:

```shell
export KIVIKDSN=http://localhost:5984/db
kivik get document SpaghettiWithMeatballs
kivik get http://localhost:5984/db/SpaghettiWithMeatballs
```

It should now be apparent why the `kivik get document` variant may be preferable for some use cases.

## Output format

`kivik` supports multiple output formats. For most commands that do not output documents, the default is intended for human consumption.

```shell
$ kivik describe doc SpaghettiWithMeatballs
      ID: SpaghettiWithMeatballs
Revision: 1-917fa2381192822767f010b95b45325b
    Size: 271
```

For documents, the default output format is formatted JSON:

```shell
$ kivik get doc SpaghettiWithMeatballs
{
        "_id": "SpaghettiWithMeatballs",
        "_rev": "1-917fa2381192822767f010b95b45325b",
        "description": "An Italian-American dish that usually consists of spaghetti, tomato sauce and meatballs.",
        "ingredients": [
                "spaghetti",
                "tomato sauce",
                "meatballs"
        ],
        "name": "Spaghetti with meatballs"
}
```

You may specify an alternative output format with the `--format` flag.  The supported output formats are:

- `json` (with optional custom indentation)
- `raw` (outputs the data exactly as provided by CouchDB)
- `yaml`
- `go-template`

### JSON

So specify custom indentation with json output, append an `=` followed by the desired indentation character(s).  Example:

```shell
$ kivik get doc SpaghettiWithMeatballs --format json=_
{
_"_id": "SpaghettiWithMeatballs",
_"_rev": "1-917fa2381192822767f010b95b45325b",
_"description": "An Italian-American dish that usually consists of spaghetti, tomato sauce and meatballs.",
_"ingredients": [
__"spaghetti",
__"tomato sauce",
__"meatballs"
_],
_"name": "Spaghetti with meatballs"
}
```

### Raw output

The raw output is likely most useful when redirecting output to a file for machine use.

```
$ kivik get doc SpaghettiWithMeatballs --format raw
{"_id":"SpaghettiWithMeatballs","_rev":"1-917fa2381192822767f010b95b45325b","description":"An Italian-American dish that usually consists of spaghetti, tomato sauce and meatballs.","ingredients":["spaghetti","tomato sauce","meatballs"],"name":"Spaghetti with meatballs"}
```
### YAML

The YAML output is pretty self explanatory:

```shell
$ kivik get doc SpaghettiWithMeatballs --format yaml
_id: SpaghettiWithMeatballs
_rev: 1-917fa2381192822767f010b95b45325b
description: An Italian-American dish that usually consists of spaghetti, tomato sauce and meatballs.
ingredients:
    - spaghetti
    - tomato sauce
    - meatballs
name: Spaghetti with meatballs
```

### Go templates

The `go-template` output format provides immense flexibility, by allowing you to specify an arbitrary output format.

Refer to the [Go template documentation](https://golang.org/pkg/text/template/) for a complete explanation of what syntax and options Go templates provide.

```shell
$ kivik get doc SpaghettiWithMeatballs --format go-template='The name is: {{ .name }}`
The name is: Spaghetti with meatballs
```

## Input

For operations which require input data, such as creating a document, you may provide the required input via the command line, or from a file (including stdin). The following commands are equivalent:

```shell
$ kivik post doc SpaghettiWithMeatballs -d '{ ... }'
$ echo '{ ... }' | kivik post doc SpaghettiWithMeatballs -D -
```

Further, by providing the `--yaml` flag, or by reading from a file with a `.yml` or `.yaml` extension, the input data is interpreted as YAML, and converted to JSON before sending to CouchDB.

## Network Controls

`kivik` provides a number of network configuratione parameters to facilitate working during network problems, or while waiting for a CouchDB server to come online.


      --request-timeout string       The time limit for each request.
      --retry int                    In case of transient error, retry up to this many times. A negative value retries forever.
      --retry-delay string           Delay between retry attempts. Disables the default exponential backoff algorithm.
      --retry-timeout string         When used with --retry, no more retries will be attempted after this timeout.

A common use case for these controls might be to wait for a new server to come up, as in this example, which will continue retrying a ping for up to 10 minutes:

```shell
$ kivik ping http://localhost:6005/ --retry-timeout=10m --retry=-1
Warning: Transient problem: Head "http://localhost:6005/_up": dial tcp [::1]:6005: connect: connection refused. Will retry in 1.03s.
Warning: Transient problem: Head "http://localhost:6005/_up": dial tcp [::1]:6005: connect: connection refused. Will retry in 1.06s.
Warning: Transient problem: Head "http://localhost:6005/_up": dial tcp [::1]:6005: connect: connection refused. Will retry in 4.22s.
...
```

## Replication

`kivik` supports two different ways of controlling replications. The `kivik post replicate` command will create a replication on a remote CouchDB server, via the `/_replicate` endpoint.

Perhaps the more interesting replication mode is `kivik`-controlled replications, which supports replication between arbitrary CouchDB servers as well as local filesystem directories.  This ability allows `kivik` to operate much like [`pg_dump`](https://www.postgresql.org/docs/12/app-pgdump.html) and [`pg_restore`](https://www.postgresql.org/docs/12/app-pgrestore.html).

`kivik`-controlled replications implement a subset of the complete CouchDB replication protocol. The main limitation is that it does not track replication state, so every replication must start from the beginning. This makes it suitable only for small replications, for example when bootstrapping a new CouchDB server, which is the primary intent of this feature.

The filesystem support is implemented using the [Kivik filesystem driver](https://github.com/go-kivik/fsdb). This means that if you replicate a CouchDB server to a local filesystem, you can interact with that filesystem via the [kivik Go client interface](https://github.com/go-kivik/kivik).

For example, to replicate from a CouchDB server on localhost, to a local directory:

```shell
$ kivik replicate -O source=http://localhost:5984/foo -O target=./dump
{
        "doc_write_failures": 0,
        "docs_read": 1,
        "docs_written": 1,
        "end_time": "2021-04-27T21:36:17.196025894+02:00",
        "missing_checked": 1,
        "missing_found": 1,
        "start_time": "2021-04-27T21:36:17.180810769+02:00"
}
```

The target directory will then contain a single `.json` file for each document that existed on the CouchDB server:

```shell
$ ls -la ./dump
$ ls -la ./asdf/
total 12
drwxr-xr-x  2 jonhall jonhall 4096 Apr 27 21:39 .
drwxr-xr-x 11 jonhall jonhall 4096 Apr 27 21:35 ..
-rw-------  1 jonhall jonhall  308 Apr 27 21:36 SpaghettiWithMeatballs.json
```

When replicating from a filesystem directory to a remote CouchDB server, `kivik` also understands YAML files, if they have a `.yml` or `.yaml` extension, to facilitate human editing of files, such as may be stored in version control.
