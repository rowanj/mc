package mcproto

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
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

func (p *packetData) String() string {
	count := len(p.data)
	switch {
	case count == 0:
		return fmt.Sprintf("PacketData(%v, %v bytes)", p.packetId, count)
	case count <= 20:
		return fmt.Sprintf("PacketData(%v, %v bytes: [%v])", p.packetId, count, hex.EncodeToString(p.data[:count]))
	}
	return fmt.Sprintf("PacketData(%v, %v bytes: [%v...])", p.packetId, count, hex.EncodeToString(p.data[:20]))
}

func EncodeTo(w io.Writer, p PacketData) error {
	var buf bytes.Buffer

	_, encodeIdErr := buf.Write(proto.EncodeVarint(p.PacketId()))
	if encodeIdErr != nil {
		return encodeIdErr
	}
	_, writeErr := buf.Write(p.Data())
	if writeErr != nil {
		return writeErr
	}

	return WriteLengthPrefixedTo(w, buf.Bytes())
}

func ReadPacket(source *bytes.Buffer) PacketData {
	payload, bytesRead := ReadLengthPrefixedFrom(source)
	if bytesRead == 0 {
		return nil
	}

	id, idLen := proto.DecodeVarint(payload)
	if idLen <= 0 {
		return nil
	}

	return &packetData{
		packetId: id,
		data:     payload[idLen:],
	}
}

func InflatePacket(source []byte) (PacketData, int) {
	payloadLen, payloadLenLen := proto.DecodeVarint(source)
	if payloadLenLen <= 0 {
		return nil, 0
	}
	end := payloadLenLen + int(payloadLen)
	if end > len(source) {
		return nil, 0
	}

	dataLen, dataLenLen := proto.DecodeVarint(source)
	if dataLenLen <= 0 {
		return nil, 0
	}

	blobStart := payloadLenLen + dataLenLen

	var packetId uint64
	var data []byte
	if dataLen == 0 {
		// packet is actually uncompressed
		dataLen = payloadLen - uint64(dataLenLen)

		var packetIdLen int
		packetId, packetIdLen = proto.DecodeVarint(source[blobStart:end])
		if packetIdLen <= 0 {
			return nil, 0
		}
		data = source[blobStart+packetIdLen : end]
	} else {
		r, err := zlib.NewReader(bytes.NewReader(data[blobStart:end]))
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
