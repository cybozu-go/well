[![GoDoc](https://godoc.org/github.com/cybozu-go/cmd?status.svg)][godoc]
[![Build Status](https://travis-ci.org/cybozu-go/cmd.svg?branch=master)](https://travis-ci.org/cybozu-go/cmd)
[![Go Report Card](https://goreportcard.com/badge/github.com/cybozu-go/cmd)](https://goreportcard.com/report/github.com/cybozu-go/cmd)
[![License](https://img.shields.io/github/license/cybozu-go/cmd.svg?maxAge=2592000)](LICENSE)

Go Command Framework
====================

This is a command framework mainly for our Go products.

Be warned that this is a _framework_ rather than a library.
Most features cannot be configured.

Features
--------

* Logging options.
* [Context](https://golang.org/pkg/context/)-based goroutine management.
* Signal handlers.
* Graceful stop for network servers.
* Enhanced [http.Server](https://golang.org/pkg/net/http/#Server).

Requirements
------------

Go 1.7 or better.

Specifications
--------------

Commands using this framework implement these features:

### Command-line options

* `-logfile FILE`

    Output logs to FILE instead of standard error.

* `-loglevel LEVEL`

    Change logging threshold to LEVEL.  Default is `info`.  
    LEVEL is one of `critical`, `error`, `warning`, `info`, or `debug`.

* `-logfmt FORMAT`

    Change log formatter.  Default is `plain`.  
    FORMAT is one of `plain`, `logfmt`, or `json`.

### Signal Handlers

* `SIGUSR1`

    If `-logfile` is specified, this signal make the program reopen
    the log file to cooperate with an external log rotation program.

    On Windows, this is not implemented.

* `SIGINT` and `SIGTERM`

    These signals cancel the context and hence goroutines and
    gracefully stop TCP and HTTP servers if any, then terminate
    the program.

    On Windows, only `SIGINT` is handled.

Usage
-----

Read [the documentation][godoc].

License
-------

[MIT][]

[godoc]: https://godoc.org/github.com/cybozu-go/cmd
[MIT]: https://opensource.org/licenses/MIT
