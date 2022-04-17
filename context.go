package telnetd

import (
	"context"
	// "encoding/hex"
	"net"
	"sync"
)

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation.
type contextKey struct {
	name string
}

var (
	// ContextKeyUser is a context key for use with Contexts in this package.
	// The associated value will be of type string.
	ContextKeyUser = &contextKey{"user"}

	// ContextKeySessionID is a context key for use with Contexts in this package.
	// The associated value will be of type string.
	ContextKeySessionID = &contextKey{"session-id"}

	// ContextKeyPermissions is a context key for use with Contexts in this package.
	// The associated value will be of type *Permissions.
	ContextKeyPermissions = &contextKey{"permissions"}

	// ContextKeyClientVersion is a context key for use with Contexts in this package.
	// The associated value will be of type string.
	ContextKeyClientVersion = &contextKey{"client-version"}

	// ContextKeyServerVersion is a context key for use with Contexts in this package.
	// The associated value will be of type string.
	ContextKeyServerVersion = &contextKey{"server-version"}

	// ContextKeyLocalAddr is a context key for use with Contexts in this package.
	// The associated value will be of type net.Addr.
	ContextKeyLocalAddr = &contextKey{"local-addr"}

	// ContextKeyRemoteAddr is a context key for use with Contexts in this package.
	// The associated value will be of type net.Addr.
	ContextKeyRemoteAddr = &contextKey{"remote-addr"}

	// ContextKeyServer is a context key for use with Contexts in this package.
	// The associated value will be of type *Server.
	ContextKeyServer = &contextKey{"ssh-server"}

	// ContextKeyConn is a context key for use with Contexts in this package.
	// The associated value will be of type gossh.ServerConn.
	ContextKeyConn = &contextKey{"telnet-conn"}

	// ContextKeyPublicKey is a context key for use with Contexts in this package.
	// The associated value will be of type PublicKey.
	ContextKeyPublicKey = &contextKey{"public-key"}
)

// Context is a package specific context interface. It exposes connection
// metadata and allows new values to be easily written to it. It's used in
// authentication handlers and callbacks, and its underlying context.Context is
// exposed on Session in the session Handler. A connection-scoped lock is also
// embedded in the context to make it easier to limit operations per-connection.
type Context interface {
	context.Context
	sync.Locker

	// User returns the username used when establishing the SSH connection.
	User() string

	// SessionID returns the session hash.
	SessionID() string

	// ClientVersion returns the version reported by the client.
	ClientVersion() string

	// ServerVersion returns the version reported by the server.
	ServerVersion() string

	// RemoteAddr returns the remote address for this connection.
	RemoteAddr() net.Addr

	// LocalAddr returns the local address for this connection.
	LocalAddr() net.Addr

	// Permissions returns the Permissions object used for this connection.
	// Permissions() *Permissions

	// SetValue allows you to easily write new values into the underlying context.
	SetValue(key, value interface{})
}

type telnetContext struct {
	context.Context
	*sync.Mutex
}

func newContext(srv *Server) (*telnetContext, context.CancelFunc) {
	innerCtx, cancel := context.WithCancel(context.Background())
	ctx := &telnetContext{innerCtx, &sync.Mutex{}}
	ctx.SetValue(ContextKeyServer, srv)
	return ctx, cancel
}

func (ctx *telnetContext) SetValue(key, value interface{}) {
	ctx.Context = context.WithValue(ctx.Context, key, value)
}

func (ctx *telnetContext) User() string {
	return ctx.Value(ContextKeyUser).(string)
}

func (ctx *telnetContext) SessionID() string {
	return ctx.Value(ContextKeySessionID).(string)
}

func (ctx *telnetContext) ClientVersion() string {
	return ctx.Value(ContextKeyClientVersion).(string)
}

func (ctx *telnetContext) ServerVersion() string {
	return ctx.Value(ContextKeyServerVersion).(string)
}

func (ctx *telnetContext) RemoteAddr() net.Addr {
	if addr, ok := ctx.Value(ContextKeyRemoteAddr).(net.Addr); ok {
		return addr
	}
	return nil
}

func (ctx *telnetContext) LocalAddr() net.Addr {
	return ctx.Value(ContextKeyLocalAddr).(net.Addr)
}
