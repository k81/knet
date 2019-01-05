package main

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/k81/knet"
)

const (
	keyHeader = "key_header"
)

type StreamHeader struct {
	Caller string
}

type StreamData struct {
	Data []byte
}

type AudioProtocol struct{}

func (p *AudioProtocol) Decode(session *knet.IoSession, reader io.Reader) (knet.Message, error) {
	_, ok := session.GetAttr(keyHeader).(*StreamHeader)
	if !ok {
		var callerLen uint16
		err := binary.Read(reader, binary.BigEndian, &callerLen)
		if err != nil {
			return nil, err
		}

		buf := &bytes.Buffer{}
		if _, err = io.CopyN(buf, reader, int64(callerLen)); err != nil {
			return nil, err
		}

		header := &StreamHeader{
			Caller: buf.String(),
		}
		session.SetAttr(keyHeader, header)
		return header, nil
	}

	buf := &bytes.Buffer{}
	if _, err := io.CopyN(buf, reader, 512); err != nil {
		return nil, err
	}

	data := &StreamData{
		Data: buf.Bytes(),
	}

	return data, nil
}

func (p *AudioProtocol) Encode(session *knet.IoSession, m knet.Message) (data []byte, err error) {
	return nil, nil
}
