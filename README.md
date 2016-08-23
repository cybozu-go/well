[![GitHub release](https://img.shields.io/github/release/cybozu-go/cmd.svg?maxAge=60)][releases]
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
* Activity tracking.

Requirements
------------

Go 1.7 or better.

Specifications
--------------

Commands using this framework implement these external specifications:

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

    These signals cancel the context of the global environment,
    and hence goroutines registered with the environment.  Usually
    this will result in graceful stop of network servers, if any.

    On Windows, only `SIGINT` is handled.

### Environment variables

* `REQUEST_ID_HEADER`

    The value of this variable is used as HTTP header name.
    The HTTP header is used to track activities across services.
    The default header name is "X-Cybozu-Request-ID".

Usage
-----

Read [Tutorial][wiki], [the design notes](DESIGN.md) and [godoc][].

License
-------

[MIT][]

[releases]: https://github.com/cybozu-go/cmd/releases
[godoc]: https://godoc.org/github.com/cybozu-go/cmd
[wiki]: https://github.com/cybozu-go/cmd/wiki/Tutorial
[MIT]: https://opensource.org/licenses/MIT
