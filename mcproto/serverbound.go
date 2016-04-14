package mcproto

import (
	"bytes"
	"fmt"
)

type SB struct{}

type SBHandshake00 struct {
	ProtocolVersion uint64
	ServerAddress   string
	ServerPort      uint16
	NextState       uint64
}

func (s *SBHandshake00) PacketId() uint64 { return 0 }
func (s *SBHandshake00) Data() []byte {
	return nil
}

func (s SB) DecodeHandshake(p PacketData) (PacketData, error) {
	if p.PacketId() == 0x00 {
		var buf bytes.Buffer
		buf.Write(p.Data())

		protocolVersion, err := ReadVarintFrom(&buf)
		if err != nil {
			return nil, err
		}
		serverAddress, err := ReadStringFrom(&buf)
		if err != nil {
			return nil, err
		}
		serverPort, err := ReadUshortFrom(&buf)
		if err != nil {
			return nil, err
		}
		nextState, err := ReadVarintFrom(&buf)
		if err != nil {
			return nil, err
		}
		return &SBHandshake00{
			ProtocolVersion: protocolVersion,
			ServerAddress:   serverAddress,
			ServerPort:      serverPort,
			NextState:       nextState,
		}, nil
	}
	return nil, fmt.Errorf("handshake error decoding %v", p)
}
