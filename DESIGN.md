Design notes
============

Logging
-------

The framework uses [cybozu-go/log][log] for structured logging, and
provides command-line flags to configure logging.

If a command-line flag is used to write logs to an external file, the
framework installs SIGUSR1 signal handler to reopen the file to work with
external log rotation programs.

HTTP Server
-----------

We extend [http.Server](https://golang.org/pkg/net/http/#Server) for:

1. Graceful server termination

    `http.Server.ConnState` callback can be used to track active
    connections.  By counting active connections (until it becomes zero),
    we can gracefully stop the server.

2. Better logging

    Use [cybozu-go/log][log] for structured
    logging of error messages, and output access logs by wrapping
    `http.Server.Handler`.

3. Cancel running handlers

    Since Go 1.7, http.Request has Context() that returns a context
    that will be canceled when Handler.ServeHTTP() returns.  The framework
    replaces the context so that the context is also canceled when the
    server is about to stop in addition to the original behavior.

To implement these, the framework provides this function:

* `HTTPServer(serv *http.Server, logger *log.Logger) *http.Server`

    This function makes a shallow copy of serv and replaces `Handler`,
    `ConnState`, and `ErrorLog` with functions that implement
    aforementioned specifications.  The `Handler` in `serv` will be
    wrapped and hence must be non-nil.

    If `logger` is non-nil, it is used to output access logs.
    If `logger` is nil, the default logger is used.
    Note that `logger` is a [cybozu-go/log][log]'s [Logger](https://godoc.org/github.com/cybozu-go/log#Logger).

Context
-------

The framework creates a single **base context** that will be canceled
upon signals and/or errors.

Goroutine management
--------------------

The framework provides a few functions to manage goroutines:

* `Stop()`

    This function cancels the base context and closes all managed
    listeners.  Once `Stop()` is called, calls for `Go()` and
    `Serve()` will fail with non-nil error.

* `Go(f func(ctx context.Context) error) error`

    This function starts a goroutine that executes `f`.  If `f` returns
    non-nil error, the framework calls `Stop()`.
    `ctx` is the base context.

* `Serve(l net.Listener, s Server) error`

    This function adds `l` to the list of listeners to be managed, then
    call `s.Serve(l)` in a managed goroutine.  `Server` is an interface
    that consists of just `Serve(l net.Listener) error` method.

* `Wait() error`

    This function waits for SIGINT or SIGTERM signal.
    If such a signal is received, it calls `Stop()` and waits all managed
    goroutines to finish.


[log]: https://github.com/cybozu-go/log/
