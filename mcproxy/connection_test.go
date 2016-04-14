package mcproxy

import (
	"net"
	"testing"
)

const testBindAddr string = "127.0.0.1:25566"
const testOriginAddr string = "10.0.3.2:25565"

func initTestServer(t *testing.T) *Server {
	laddr, addrErr := net.ResolveTCPAddr("tcp", testBindAddr)
	if addrErr != nil {
		t.Fatal(addrErr)
	}
	s, serverErr := NewServer(testOriginAddr, laddr)
	if serverErr != nil {
		t.Fatal(serverErr)
	}
	return s
}

func TestServerInit(t *testing.T) {
	s := initTestServer(t)
	defer s.Close()
}

func TestServerConnect(t *testing.T) {
	s := initTestServer(t)
	defer s.Close()

	go func() {
		e := s.Run()
		if e != nil {
			t.Fatal(e)
		}
	}()

	c, err := net.DialTCP("tcp", nil, s.originAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
}
