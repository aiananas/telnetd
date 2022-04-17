package main

import (
	"fmt"
	"net"

	"telnetd"

	"github.com/pires/go-proxyproto"
)

func main() {
	ln, err := net.Listen("tcp", ":10023")
	if err != nil {
		fmt.Errorf(err.Error())
		return
	}
	proxyListener := &proxyproto.Listener{Listener: ln}

	var srv telnetd.Server

	if err := srv.Serve(proxyListener); err != nil {
		fmt.Errorf(err.Error())
		return
	}

	// telnetd
	// err :=
	// logger.Fatal(s.Srv.Serve(proxyListener))
}
