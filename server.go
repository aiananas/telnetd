package telnetd

import (
	// "context"
	"errors"
	// "fmt"
	"net"
	"sync"
	"time"
)

var ErrServerClosed = errors.New("telnet: Server closed")

type Server struct {
	Addr    string  // TCP address to listen on, ":23" if empty
	Handler Handler // handler to invoke, ssh.DefaultHandler if nil

	ConnCallback             ConnCallback
	ConnectionFailedCallback ConnectionFailedCallback

	IdleTimeout time.Duration // connection timeout when no activity, none if empty
	MaxTimeout  time.Duration // absolute connection timeout, none if empty

	listenerWg sync.WaitGroup
	mu         sync.RWMutex
	listeners  map[net.Listener]struct{}
	conns      map[*net.Conn]struct{}
	connWg     sync.WaitGroup
	doneChan   chan struct{}
}

func (srv *Server) getDoneChan() <-chan struct{} {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	return srv.getDoneChanLocked()
}

func (srv *Server) getDoneChanLocked() chan struct{} {
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	return srv.doneChan
}

func (srv *Server) closeDoneChanLocked() {
	ch := srv.getDoneChanLocked()
	select {
	case <-ch:
		// Already closed. Don't close again.
	default:
		// Safe to close here. We're the only closer, guarded
		// by srv.mu.
		close(ch)
	}
}

func (srv *Server) Serve(l net.Listener) error {
	defer l.Close()

	if srv.Handler == nil {
		srv.Handler = DefaultHandler
	}
	var tempDelay time.Duration

	srv.trackListener(l, true)
	defer srv.trackListener(l, false)
	for {
		conn, e := l.Accept()
		if e != nil {
			select {
			case <-srv.getDoneChan():
				return ErrServerClosed
			default:
			}
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		go srv.HandleConn(conn)
	}
}

func (srv *Server) HandleConn(newConn net.Conn) {
	ctx, cancel := newContext(srv)
	if srv.ConnCallback != nil {
		cbConn := srv.ConnCallback(ctx, newConn)
		if cbConn == nil {
			newConn.Close()
			return
		}
		newConn = cbConn
	}
	conn := &serverConn{
		Conn:          newConn,
		idleTimeout:   srv.IdleTimeout,
		closeCanceler: cancel,
	}
	if srv.MaxTimeout > 0 {
		conn.maxDeadline = time.Now().Add(srv.MaxTimeout)
	}

	ctx.SetValue(ContextKeyConn, conn)

	// 需要持续性处理数据
}

func (srv *Server) trackListener(ln net.Listener, add bool) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if srv.listeners == nil {
		srv.listeners = make(map[net.Listener]struct{})
	}
	if add {
		// If the *Server is being reused after a previous
		// Close or Shutdown, reset its doneChan:
		if len(srv.listeners) == 0 && len(srv.conns) == 0 {
			srv.doneChan = nil
		}
		srv.listeners[ln] = struct{}{}
		srv.listenerWg.Add(1)
	} else {
		delete(srv.listeners, ln)
		srv.listenerWg.Done()
	}
}

// SetOption runs a functional option against the server.
func (srv *Server) SetOption(option Option) error {
	// NOTE: there is a potential race here for any option that doesn't call an
	// internal method. We can't actually lock here because if something calls
	// (as an example) AddHostKey, it will deadlock.

	//srv.mu.Lock()
	//defer srv.mu.Unlock()

	return option(srv)
}

// func (srv *Server) trackConn(c *gossh.ServerConn, add bool) {
// 	srv.mu.Lock()
// 	defer srv.mu.Unlock()

// 	if srv.conns == nil {
// 		srv.conns = make(map[*gossh.ServerConn]struct{})
// 	}
// 	if add {
// 		srv.conns[c] = struct{}{}
// 		srv.connWg.Add(1)
// 	} else {
// 		delete(srv.conns, c)
// 		srv.connWg.Done()
// 	}
// }
