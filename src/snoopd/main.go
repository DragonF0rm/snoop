package main

import (
	"io"
	"net"
	"os"
	"snoopd/cfg"
	"snoopd/log"
	"snoopd/proxy"
)

const SockAddr = "/tmp/snoopd.sock"

func echoServer(c net.Conn) {
	//log.Printf("Client connected [%s]", c.RemoteAddr().Network())
	io.Copy(c, c)
	c.Close()
}

func main() {
	serverName := cfg.GetString("snoopd.name")
	serverVersion := cfg.GetString("snoopd.version")
	log.Info("Starting", serverName, serverVersion)

	go proxy.ListenAndServe()
	go proxy.ListenAndServeTLS()

	if err := os.RemoveAll(SockAddr); err != nil {
		log.Fatal("Unable to remove SockAddr <" + SockAddr + ">:", err)
	}

	l, err := net.Listen("unix", SockAddr)
	if err != nil {
		log.Fatal("Unable to listen:", err)
	}
	defer l.Close()

	for {
		// Accept new connections, dispatching them to echoServer
		// in a goroutine.
		conn, err := l.Accept()
		if err != nil {
			//TODO fixme
			log.Fatal("accept error:", err)
		}

		go echoServer(conn)
	}
}
