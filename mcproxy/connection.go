package mcproxy

import (
	"bytes"
	"log"
	"net"
	"sync"

	"github.com/rowanj/mcproto"
)

const bufferSize = 65536

type connectionState byte

const (
	handshaking connectionState = iota
	status
	login
	play
)

type connection struct {
	state  connectionState
	client *net.TCPConn
	origin *net.TCPConn
}

func NewConnection(client *net.TCPConn, origin *net.TCPConn) *connection {

	client.SetReadBuffer(bufferSize)
	origin.SetReadBuffer(bufferSize)

	log.Printf("new connection from %v", client.RemoteAddr())
	return &connection{
		state:  handshaking,
		client: client,
		origin: origin,
	}
}

func (c *connection) Run() {
	var wg sync.WaitGroup
	go c.handleMessageStream(wg)
	go c.handleMessageStream(wg)
	wg.Wait()
}

func (c *connection) Close() {
	c.client.Close()
	c.origin.Close()
}

func (c *connection) handleMessageStream(wg sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	var buf bytes.Buffer
	var raw []byte = make([]byte, bufferSize)
	for {
		n, readErr := c.client.Read(raw)
		if n == 0 {
			return
		}
		log.Printf("received %v bytes from %v", n, c.client.RemoteAddr())
		buf.Write(raw[:n])

		p, packetLen := mcproto.DecodePacket(buf.Bytes())
		if p != nil {
			log.Printf("read packet ID %v: %v bytes", p.PacketId(), packetLen)
			written, writeErr := c.origin.Write(buf.Next(int(packetLen)))
			if uint64(written) != packetLen {
				log.Printf("sent %v/%v for %v (%v)", written, packetLen, c.client.RemoteAddr(), writeErr)
			}
		} else {
			log.Printf("warn: fragment from %v (%v bytes)", c.client.RemoteAddr(), n)
		}

		if readErr != nil {
			log.Printf("error: %v - %v", c.client.RemoteAddr(), readErr)
			return
		}
	}
}
