package main

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/k81/knet"
)

const (
	PktTypeHandshake = 1
	PktTypeAudioData = 2
)

type Header struct {
	Type uint32
	Len  uint32
}

type Packet struct {
	Header
	Data []byte
}

const (
	keyReadBuf = "key_readbuf"
)

type AudioProtocol struct{}

func (p *AudioProtocol) Decode(session *knet.IoSession, reader io.Reader) (knet.Message, error) {
	pkt := &Packet{}
	err := binary.Read(reader, binary.BigEndian, &pkt.Header)
	if err != nil {
		return nil, err
	}

	readBuf, ok := session.GetAttr(keyReadBuf).(*bytes.Buffer)
	if !ok {
		readBuf = &bytes.Buffer{}
		session.SetAttr(keyReadBuf, readBuf)
	}
	readBuf.Reset()

	if _, err = io.CopyN(readBuf, reader, int64(pkt.Len)); err != nil {
		return nil, err
	}

	pkt.Data = readBuf.Bytes()

	return pkt, nil
}

func (p *AudioProtocol) Encode(session *knet.IoSession, m knet.Message) (data []byte, err error) {
	return nil, nil
}
