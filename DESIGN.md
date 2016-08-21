Design notes
============

Logging
-------

The framework uses [cybozu-go/log][log] for structured logging, and
provides command-line flags to configure logging.

The framework also provides `LogConfig` struct that can be load from
JSON or [TOML][] file, and configures the default logger according to
the struct member values.  The command-line flags take precedence
over the member values, if specified.

Context
-------

The framework creates a single **base context** that will be canceled
when `Stop()` is called.  `Stop()` is described later.

Goroutine management
--------------------

The framework provides following functions to manage goroutines:

* `Go(f func(ctx context.Context) error)`

    This function starts a goroutine that executes `f`.  If `f` returns
    non-nil error, the framework calls `Cancel()` with that error.

    `ctx` is a derived context from the base context that is to be
    canceled when f returns.

* `Stop()`

    This function just declares no further `Go()` will be called.

    Calling `Stop()` is optional if and only if `Cancel()` is
    guaranteed to be called at some point.  For instance, if the
    program runs until SIGINT or SIGTERM, `Stop()` is optional.

* `Cancel(err error)`

    This function cancels the base context and closes all managed
    listeners.  After `Cancel()`, `Go()` will not start new goroutines
    any longer.

* `Wait() error`

    This function waits for `Stop()` or `Cancel()` being called and then
    waits for all managed goroutines to finish.  The return value will be
    the error that was passed to `Cancel()`, or nil.

Signal handlers
---------------

The framework implicitly starts a goroutine to handle SIGINT and SIGTERM.
The goroutine, when such a signal is sent, will call `Cancel()` with an
error indicating SIGINT or SIGTERM is got.

If a command-line flag is used to write logs to an external file, the
framework starts SIGUSR1 signal handler to reopen the file to work with
external log rotation programs.

Related functions:

* `IsSignaled(err error) bool`

    This function returns `true` if `err` returned from `Wait()` is
    a result of SIGINT/SIGTERM handlers.

Generic server
--------------

Suppose that we create a simple TCP server on this framework.

A naive idea is to use `Go()` to start goroutines for every accepted
connections.  However, since `Go()` acquires a package global mutex,
such an implementation would limit concurrency of the server.

In order to implement high performance servers, the server should
manage all goroutines started by the server by itself.  The framework
provides `Server` as a generic implementation of such servers.

HTTP Server
-----------

As to [http.Server](https://golang.org/pkg/net/http/#Server), we extend it for:

1. Graceful server termination

    `http.Server` can gracefully shutdown by the following steps:

    1. Close all listeners.
    2. Call `SetKeepAlivesEnabled(false)`.
    3. Call `net.Conn.SetReadDeadline(time.Now())` for all idle connections.
    4. Wait all connections to be closed.

    For 3 and 4, `http.Server.ConnState` callback can be used to track
    connection status.

    Note that `ReadTimeout` can work as keep-alive timeout according
    to the go source code (at least as of Go 1.7).  Since keep-alived
    connections may block on `conn.Read` to wait for the next request,
    we need to cancel `conn.Read` for quicker shutdown.

    c.f. https://groups.google.com/forum/#!topic/golang-nuts/5E4gM7EzdLw

2. Better logging

    Use [cybozu-go/log][log] for structured logging of error messages,
    and output access logs by wrapping `http.Server.Handler`.

3. Cancel running handlers

    Since Go 1.7, `http.Request` has `Context()` that returns a context
    that will be canceled when `Handler.ServeHTTP()` returns.  The
    framework replaces the context so that the context is also canceled
    when the server is about to stop in addition to the original behavior.

To implement these, the framework provides a wrapping struct:

* `HTTPServer`

    This struct embeds http.Server and overrides `Serve`, `ListenAndServe`,
    and `ListenAndServeTLS` methods.


[log]: https://github.com/cybozu-go/log/
[TOML]: https://github.com/toml-lang/toml
