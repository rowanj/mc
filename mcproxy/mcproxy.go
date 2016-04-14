package mcproxy

import (
	"log"

	"github.com/rowanj/mc/mcproto"
)

type processPacket struct {
}

func (c *connection) handlePackets() error {
	defer log.Printf("exiting handlePackets")

	err := c.handshake()
	if err != nil {
		return err
	}

	clientIn := c.clientIn
	clientOut := c.clientOut
	originIn := c.originIn
	originOut := c.originOut

	done := make(chan bool)

	forwardPackets := func(in, out chan mcproto.PacketData, dir mcproto.Direction) {
		for p := range in {
			out <- p
		}
		done <- true
	}

	go forwardPackets(clientIn, originOut, mcproto.ServerBound)
	go forwardPackets(originIn, clientOut, mcproto.ClientBound)

	<-done

	return nil
}

func (c *connection) handshake() error {
	sb := mcproto.SB{}
	handshakeRq := <-c.clientIn
	p, err := sb.DecodeHandshake(handshakeRq)
	if err != nil {
		return err
	}
	log.Printf("%+v", p)
	return nil
}
