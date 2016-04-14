package mcproxy

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"github.com/rowanj/mc/mcproto"
)

const bufferSize = 65536

type direction int

const (
	clientToServer direction = iota
	serverToClient
)

type connection struct {
	client *net.TCPConn
	origin *net.TCPConn
}

func NewConnection(client *net.TCPConn, origin *net.TCPConn) *connection {
	client.SetNoDelay(true)
	client.SetReadBuffer(bufferSize)
	origin.SetNoDelay(true)
	origin.SetReadBuffer(bufferSize)

	log.Printf("new connection from %v", client.RemoteAddr())
	return &connection{
		client: client,
		origin: origin,
	}
}

func (c *connection) Run() error {

	clientPackets := make(chan mcproto.PacketData)
	originPackets := make(chan mcproto.PacketData)

	errors := make(chan error)

	tidy := func() {
		c.origin.Close()
		c.client.Close()
	}
	go func() {
		defer tidy()
		defer close(clientPackets)
		errors <- c.handleTcpStream(clientToServer, clientPackets)
	}()
	go func() {
		defer tidy()
		defer close(originPackets)
		errors <- c.handleTcpStream(serverToClient, originPackets)
	}()

	go func() {
		defer tidy()
		errors <- c.handlePacket(clientToServer, clientPackets)
	}()
	go func() {
		defer tidy()
		errors <- c.handlePacket(serverToClient, originPackets)
	}()

	err := <-errors

	for tasks := 4; tasks > 1; tasks-- {
		<-errors
	}

	return err
}

func (c *connection) Close() {
	c.client.Close()
	c.origin.Close()
}

func (c *connection) handleTcpStream(dir direction, rx chan mcproto.PacketData) error {

	var source *net.TCPConn
	var idString string

	if dir == clientToServer {
		source = c.client
		idString = fmt.Sprintf("<= %v", c.client.RemoteAddr())
	}
	if dir == serverToClient {
		source = c.origin
		idString = fmt.Sprintf("=> %v", c.client.RemoteAddr())
	}

	defer log.Printf("%v exiting", idString)

	var buf bytes.Buffer
	var raw []byte = make([]byte, bufferSize)
	for {
		n, readErr := source.Read(raw)
		if n == 0 {
			return fmt.Errorf("%v connection reset by peer", idString)
		}
		// log.Printf("%v %v bytes", idString, n)
		buf.Write(raw[:n])

		for buf.Len() > 0 {
			p := mcproto.ReadPacket(&buf)
			if buf.Len() > bufferSize {
				return fmt.Errorf("%v error: too much invalid data", idString)

			}
			if p == nil {
				break
			}
			log.Printf("%v rx %v", idString, p)
			rx <- p
		}

		if readErr != nil {
			return fmt.Errorf("%v error: %v", idString, readErr)
		}
	}
}

func (c *connection) handlePacket(dir direction, stream chan mcproto.PacketData) error {
	defer log.Printf("leaving handlePacket %v", c.client.RemoteAddr())
	var sink *net.TCPConn
	if dir == clientToServer {
		sink = c.origin
	}
	if dir == serverToClient {
		sink = c.client
	}
	for p := range stream {
		err := mcproto.EncodeTo(sink, p)
		if err != nil {
			return err
		}
		//log.Printf("forwarded %v", p)
	}
	return nil
}
