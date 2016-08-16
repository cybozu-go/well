Design notes
============

Logging
-------

The framework uses [cybozu-go/log][log] for structured logging, and
provides command-line flags to configure logging.

The framework also provides `LogConfig` struct that is initialized
by the command-line flags through `NewLogConfig` function, and has
`Apply()` method to apply logging configurations.

Since `LogConfig` is annotated with struct tags for JSON and TOML,
users may load configurations from JSON or TOML file.

HTTP Server
-----------

We extend [http.Server](https://golang.org/pkg/net/http/#Server) for:

1. Graceful server termination

    `http.Server.ConnState` callback can be used to track active
    connections.  By counting active connections (until it becomes zero),
    we can gracefully stop the server.

    Unfortunately, `ConnState` callback cannot determine whether a
    connection was `StateActive` or not when the state changes to
    `StateClosed`.  We need to implement a tracking logic for connection
    statuses.

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

* `HTTPServer(serv *http.Server, al *log.Logger) *http.Server`

    This function makes a shallow copy of serv and replaces `Handler`,
    `ConnState`, and `ErrorLog` with functions that implement
    aforementioned specifications.  The `Handler` in `serv` will be
    wrapped and hence must be non-nil.

    If `al` is non-nil, it is used to output access logs.
    If `al` is nil, the default logger is used.

Context
-------

The framework creates a single **base context** that will be canceled
when `Stop()` is called.  `Stop()` is described later.

Signal handlers
---------------

The framework implicitly starts a goroutine to handle SIGINT and SIGTERM.
The goroutine, when such a signal is sent, will call `Stop()` described
below.

If a command-line flag is used to write logs to an external file, the
framework starts SIGUSR1 signal handler to reopen the file to work with
external log rotation programs.

Related functions:

* `IsSignaled() bool`

    This function returns `true` if `Stop()` was called by SIGINT/SIGTERM
    signal handlers.

Goroutine management
--------------------

The framework provides following functions to manage goroutines:

* `Stop(err error)`

    This function cancels the base context and closes all managed
    listeners.  Once `Stop()` is called, calls for `Go()` and
    `Serve()` will fail with non-nil error.

* `Go(f func(ctx context.Context) error) error`

    This function starts a goroutine that executes `f`.  If `f` returns
    non-nil error, the framework calls `Stop()` with that error.
    `ctx` is the base context.

* `Serve(l net.Listener, s Server) error`

    This function adds `l` to the list of listeners to be managed, then
    call `s.Serve(l)` in a managed goroutine.  `Server` is an interface
    that consists of just `Serve(l net.Listener) error` method.

* `Wait() error`

    This function waits for `Stop()` being called and then waits for
    all managed goroutines to finish.  The return value will be the error
    that was passed to `Stop()`.


[log]: https://github.com/cybozu-go/log/
