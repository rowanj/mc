package mcproto

import (
	"bytes"
	"encoding/binary"
	"io"
	"unicode/utf8"

	"github.com/golang/protobuf/proto"
)

type DecodeError struct{}

func (DecodeError) Error() string {
	return "decode error"
}

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

func ReadVarintFrom(b *bytes.Buffer) (uint64, error) {
	result, resultLen := proto.DecodeVarint(b.Bytes())
	if resultLen < 1 {
		return 0, DecodeError{}
	}
	b.Next(resultLen)

	return result, nil
}

func ReadUshortFrom(b *bytes.Buffer) (uint16, error) {
	data := b.Next(2)
	result := binary.BigEndian.Uint16(data)
	return result, nil
}

func ReadStringFrom(b *bytes.Buffer) (string, error) {
	strLen, lenErr := ReadVarintFrom(b)
	if lenErr != nil {
		return "", lenErr
	}

	strData := b.Next(int(strLen))
	if !utf8.Valid(strData) {
		return "", DecodeError{}
	}

	return string(strData), nil
}
