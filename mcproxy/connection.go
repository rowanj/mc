package mcproxy

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"github.com/rowanj/mc/mcproto"
)

const bufferSize = 65536

type connection struct {
	client *net.TCPConn

	clientIn  chan mcproto.PacketData
	clientOut chan mcproto.PacketData

	origin *net.TCPConn

	originIn  chan mcproto.PacketData
	originOut chan mcproto.PacketData
}

func NewConnection(client *net.TCPConn, origin *net.TCPConn) *connection {
	setupConnection := func(c *net.TCPConn) {
		c.SetNoDelay(true)
		c.SetReadBuffer(bufferSize)
	}
	setupConnection(client)
	setupConnection(origin)

	return &connection{
		client:    client,
		clientIn:  make(chan mcproto.PacketData),
		clientOut: make(chan mcproto.PacketData),

		origin:    origin,
		originIn:  make(chan mcproto.PacketData),
		originOut: make(chan mcproto.PacketData),
	}
}

func (c *connection) Run() error {

	errors := make(chan error)

	clientAddr := c.client.RemoteAddr()

	// read incoming packets from network
	packetReader := func(r *net.TCPConn, queue chan mcproto.PacketData, idString string) {
		defer c.Close()
		defer close(queue)
		errors <- c.readTcpPackets(r, queue, idString)
	}
	go packetReader(c.client, c.clientIn, fmt.Sprintf("<- %v", clientAddr))
	go packetReader(c.origin, c.originIn, fmt.Sprintf("-> %v", clientAddr))

	// handle read packets
	go func() {
		defer close(c.clientOut)
		defer close(c.originOut)
		errors <- c.handlePackets()
	}()

	// write packets to network
	packetWriter := func(queue chan mcproto.PacketData, sink *net.TCPConn, idString string) {
		errors <- c.writeTcpPackets(queue, sink, idString)
	}
	go packetWriter(c.originOut, c.origin, fmt.Sprintf("<= %v", clientAddr))
	go packetWriter(c.clientOut, c.client, fmt.Sprintf("=> %v", clientAddr))

	err := <-errors
	// sync up
	for tasks := 5; tasks > 1; tasks-- {
		<-errors
	}

	return err
}

func (c *connection) Close() {
	c.client.Close()
	c.origin.Close()
}

func (c *connection) readTcpPackets(source *net.TCPConn, rx chan mcproto.PacketData, idString string) error {
	defer source.Close()
	defer log.Printf("%v closed TCP", idString)

	var buf bytes.Buffer
	var raw []byte = make([]byte, bufferSize)
	for {
		n, readErr := source.Read(raw)
		if n == 0 {
			rx <- nil
			return fmt.Errorf("%v connection reset by peer", source.RemoteAddr())
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

func (c *connection) writeTcpPackets(stream chan mcproto.PacketData, sink *net.TCPConn, idString string) error {
	defer sink.Close()
	defer log.Printf("%v closed TCP", idString)

	for p := range stream {
		if p == nil {
			return nil
		}
		err := mcproto.EncodeTo(sink, p)
		if err != nil {
			return err
		}
		log.Printf("%v tx %v", idString, p)
	}
	return nil
}
