package main

import (
	"log"
	"net"

	"github.com/rowanj/mc/mcproxy"
)

func main() {
	err := start()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func start() error {
	laddr, addrErr := net.ResolveTCPAddr("tcp", ":25565")
	if addrErr != nil {
		return addrErr
	}
	s, serverErr := mcproxy.NewServer("10.0.3.2:25565", laddr)
	if serverErr != nil {
		return serverErr
	}
	return s.Run()
}
