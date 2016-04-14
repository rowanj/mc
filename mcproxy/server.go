package mcproxy

import (
	"log"
	"net"
)

type Server struct {
	origin     string
	originAddr *net.TCPAddr
	listener   *net.TCPListener
}

func NewServer(origin string, laddr *net.TCPAddr) (*Server, error) {
	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return nil, err
	}

	originAddr, lookupErr := net.ResolveTCPAddr("tcp", origin)
	if lookupErr != nil {
		return nil, err
	}

	return &Server{
		origin:     origin,
		originAddr: originAddr,
		listener:   l,
	}, nil
}

func (s *Server) Run() error {
	for {
		con, acceptErr := s.listener.AcceptTCP()
		if acceptErr != nil {
			return acceptErr
		}

		go s.handleConnection(con)
	}
}

func (s *Server) handleConnection(client *net.TCPConn) {
	origin, originErr := net.DialTCP("tcp", nil, s.originAddr)
	if originErr != nil {
		log.Printf("Cannot connect to origin server %v: %v", s.originAddr, originErr)
		client.Close()
		return
	}

	con := NewConnection(client, origin)
	runErr := con.Run()
	con.Close()

	log.Printf("%v (connection closed)", runErr)
}
