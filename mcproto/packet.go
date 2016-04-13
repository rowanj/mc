package mcproto

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
)

type PacketData interface {
	PacketId() uint64
	Data() []byte
}

type packetData struct {
	packetId uint64
	data     []byte
}

func (p *packetData) PacketId() uint64 {
	return p.packetId
}

func (p *packetData) Data() []byte {
	return p.data
}

func DecodePacket(source []byte) (PacketData, uint64) {
	packetLen, packetLenLen := proto.DecodeVarint(source)
	if packetLenLen <= 0 {
		return nil, 0
	}
	packetEnd := uint64(packetLenLen) + packetLen
	if packetEnd > uint64(len(source)) {
		return nil, 0
	}

	id, idLen := proto.DecodeVarint(source[packetLenLen:])
	if idLen <= 0 {
		return nil, 0
	}

	return &packetData{
		packetId: id,
		data:     source[packetLenLen+idLen : packetEnd],
	}, packetEnd
}

func InflatePacket(source []byte) (PacketData, uint64) {
	packetLen, packetLenLen := proto.DecodeVarint(source)
	if packetLenLen <= 0 {
		return nil, 0
	}
	end := uint64(packetLenLen) + packetLen
	if end > uint64(len(source)) {
		return nil, 0
	}

	dataLen, dataLenLen := proto.DecodeVarint(source)
	if dataLenLen <= 0 {
		return nil, 0
	}

	payloadStart := packetLenLen + dataLenLen

	var packetId uint64
	var data []byte
	if dataLen == 0 {
		// packet is actually uncompressed
		dataLen = packetLen - uint64(dataLenLen)

		var packetIdLen int
		packetId, packetIdLen = proto.DecodeVarint(source[payloadStart:end])
		if packetIdLen <= 0 {
			return nil, 0
		}
		data = source[payloadStart+packetIdLen : end]
	} else {
		r, err := zlib.NewReader(bytes.NewReader(data[payloadStart:end]))
		if err != nil {
			return nil, 0
		}
		inflatedData, inflateErr := ioutil.ReadAll(r)
		if inflateErr != nil {
			return nil, 0
		}
		if uint64(len(inflatedData)) != dataLen {
			return nil, 0
		}
		var packetIdLen int
		packetId, packetIdLen = proto.DecodeVarint(inflatedData)
		if packetIdLen <= 0 {
			return nil, 0
		}
		data = inflatedData[packetIdLen:]
	}
	return &packetData{
		packetId: packetId,
		data:     data,
	}, end
}
