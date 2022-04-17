package telnetd

import (
	// "crypto/subtle"
	"net"
)

type Signal string

// POSIX signals as listed in RFC 4254 Section 6.10.
const (
	SIGABRT Signal = "ABRT"
	SIGALRM Signal = "ALRM"
	SIGFPE  Signal = "FPE"
	SIGHUP  Signal = "HUP"
	SIGILL  Signal = "ILL"
	SIGINT  Signal = "INT"
	SIGKILL Signal = "KILL"
	SIGPIPE Signal = "PIPE"
	SIGQUIT Signal = "QUIT"
	SIGSEGV Signal = "SEGV"
	SIGTERM Signal = "TERM"
	SIGUSR1 Signal = "USR1"
	SIGUSR2 Signal = "USR2"
)

// DefaultHandler is the default Handler used by Serve.
var DefaultHandler Handler

// Option is a functional option handler for Server.
type Option func(*Server) error

// Handler is a callback for handling established SSH sessions.
type Handler func(Session)

// ConnCallback is a hook for new connections before handling.
// It allows wrapping for timeouts and limiting by returning
// the net.Conn that will be used as the underlying connection.
type ConnCallback func(ctx Context, conn net.Conn) net.Conn

// ConnectionFailedCallback is a hook for reporting failed connections
// Please note: the net.Conn is likely to be closed at this point
type ConnectionFailedCallback func(conn net.Conn, err error)

// Window represents the size of a PTY window.
type Window struct {
	Width  int
	Height int
}

// Pty represents a PTY request and configuration.
type Pty struct {
	Term   string
	Window Window
	// HELP WANTED: terminal modes!
}

// Serve accepts incoming SSH connections on the listener l, creating a new
// connection goroutine for each. The connection goroutines read requests and
// then calls handler to handle sessions. Handler is typically nil, in which
// case the DefaultHandler is used.
func Serve(l net.Listener, handler Handler, options ...Option) error {
	srv := &Server{Handler: handler}
	for _, option := range options {
		if err := srv.SetOption(option); err != nil {
			return err
		}
	}
	return srv.Serve(l)
}

// // ListenAndServe listens on the TCP network address addr and then calls Serve
// // with handler to handle sessions on incoming connections. Handler is typically
// // nil, in which case the DefaultHandler is used.
// func ListenAndServe(addr string, handler Handler, options ...Option) error {
// 	srv := &Server{Addr: addr, Handler: handler}
// 	for _, option := range options {
// 		if err := srv.SetOption(option); err != nil {
// 			return err
// 		}
// 	}
// 	return srv.ListenAndServe()
// }

// Handle registers the handler as the DefaultHandler.
func Handle(handler Handler) {
	DefaultHandler = handler
}
