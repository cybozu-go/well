package cmd

import (
	"net/http"

	"github.com/cybozu-go/log"
)

// HTTPServer is a wrapper for http.Server.
//
// This struct overrides Serve and ListenAndServe* methods, and
// replaces Handler and ConnState http.Server struct members.
type HTTPServer struct {
	*http.Server

	// AccessLog is a logger for access logs.
	// If this is nil, the default logger is used.
	AccessLog *log.Logger

	// Env is the environment where this server runs.
	//
	// The global environment is used if Env is nil.
	Env *Environment
}

func wrapHTTPHandler(hdl http.Handler) http.Handler {
	return hdl
}
