package mcproto

import (
	"bytes"
	"io"

	"github.com/golang/protobuf/proto"
)

func WriteLengthPrefixedTo(w io.Writer, data []byte) error {
	var buf bytes.Buffer

	length := len(data)
	_, err := buf.Write(proto.EncodeVarint(uint64(length)))
	if err != nil {
		return err
	}

	_, err = buf.Write(data)
	if err != nil {
		return err
	}
	_, err = w.Write(buf.Bytes())
	//log.Printf("wrote %v bytes", wr)
	return err
}

func ReadLengthPrefixedFrom(b *bytes.Buffer) (data []byte, n int) {
	length, lengthLen := proto.DecodeVarint(b.Bytes())
	if lengthLen == 0 {
		return
	}
	payloadLength := int(length)
	readSize := lengthLen + payloadLength
	if readSize > b.Len() {
		// log.Printf("fragment %v/%v available", b.Len(), readSize)
		return
	}
	n = readSize
	b.Next(lengthLen)
	data = b.Next(payloadLength)
	return
}
