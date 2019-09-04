package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

const SockAddr = "/tmp/snoopd.sock"

func main() {
	addr, err := net.ResolveUnixAddr("unix", SockAddr)
	if err != nil {
		log.Fatal("Unable to resolve unix socket addr")
	}
	c,err := net.DialUnix("unix", nil, addr)
	if err != nil {
		log.Fatal("Unable to dial unix socket")
	}
	defer c.Close()

	msg := os.Args[1]
	_, err = c.Write([]byte(msg))
	if err != nil {
		log.Fatal("Unable to send message to unix socket")
	}
	buf := make([]byte, 256)
	_, err = c.Read(buf)
	if err != nil {
		log.Fatal("Unable to read message rom unix socket")
	}
	fmt.Println(string(buf))
}
