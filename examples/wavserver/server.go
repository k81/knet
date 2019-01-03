package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"strings"
	"time"

	"github.com/cryptix/wav"
	"github.com/k81/knet"
)

var (
	homeDir string
	dataDir string
)

const (
	keyCallId      = "callid"
	keyRawFileName = "raw_filename"
	keyRawFile     = "raw_file"
	keyWavFile     = "wav_file"
	keyWavEncoder  = "audio_encoder"
)

var (
	ErrInvalidPacket    = errors.New("invalid packet")
	ErrHandshakeNotDone = errors.New("handshake not done")
)

type srvAudioHandler struct {
	knet.IoHandlerAdapter
}

func (h *srvAudioHandler) OnConnected(session *knet.IoSession) error {
	log.Printf("session connected, remote_addr=%v\n", session.RemoteAddr())
	//m := echo.NewEchoMessage("welcome to the echo server, enjoy it!")
	//session.Send(m)
	return nil
}

func (h *srvAudioHandler) OnDisconnected(session *knet.IoSession) {
	log.Printf("session disconnected, remote_addr=%v, stats=%v\n", session.RemoteAddr(), session.String())
	callId, ok := session.GetAttr(keyCallId).(string)
	if !ok {
		return
	}

	rawFileName, ok := session.GetAttr(keyRawFileName).(string)
	if !ok {
		return
	}
	wavEncoder, ok := session.GetAttr(keyWavEncoder).(*wav.Writer)
	if ok {
		if err := wavEncoder.Close(); err != nil {
			log.Printf("ERROR: close wav encoder: %v\n", err)
		}
	}
	wavFile, ok := session.GetAttr(keyWavFile).(*os.File)
	if ok {
		wavFile.Close()
	}
	rawFile, ok := session.GetAttr(keyRawFile).(*os.File)
	if ok {
		rawFile.Close()
	}

	log.Printf("call finished: callid=%v, audio_filename=%v\n", callId, rawFileName)
	return
}

func (h *srvAudioHandler) OnError(session *knet.IoSession, err error) {
	log.Printf("sesson error: %v", err)
}

func (h *srvAudioHandler) OnMessage(session *knet.IoSession, m knet.Message) error {
	pkt := m.(*Packet)
	switch pkt.Type {
	case PktTypeHandshake:
		callId := string(pkt.Data)
		callId = strings.Replace(callId, "/", "_", -1)
		callId = strings.Replace(callId, "-", "_", -1)
		log.Printf("got handshake: %v\n", callId)
		session.SetAttr(keyCallId, callId)
	case PktTypeAudioData:
		callId, ok := session.GetAttr(keyCallId).(string)
		if !ok {
			log.Printf("ERROR: got audio data before handshake\n")
			return ErrHandshakeNotDone
		}

		if len(pkt.Data) == 0 {
			log.Printf("WARNING: got empty data packet: callid=%v\n", callId)
			return nil
		}

		var (
			rawFileName string
			wavFileName string
			rawFile     *os.File
			wavFile     *os.File
			wavEncoder  *wav.Writer
		)
		wavEncoder, ok = session.GetAttr(keyWavEncoder).(*wav.Writer)
		if !ok {
			rawFileName = path.Join(dataDir, fmt.Sprintf("%v_%v.raw", callId, time.Now().Format("20060102150405")))
			wavFileName = path.Join(dataDir, fmt.Sprintf("%v_%v.wav", callId, time.Now().Format("20060102150405")))

			var err error

			rawFile, err = os.Create(rawFileName)
			if err != nil {
				log.Printf("ERROR: create raw file failed: filename=%v, error=%v\n", rawFileName, err)
				return err
			}

			wavFile, err = os.Create(wavFileName)
			if err != nil {
				log.Printf("ERROR: create wav file failed: filename=%v, error=%v\n", wavFileName, err)
				return err
			}

			wavFileDesc := wav.File{
				SampleRate:      8000,
				SignificantBits: 16,
				Channels:        1,
			}
			wavEncoder, err = wavFileDesc.NewWriter(wavFile)
			if err != nil {
				log.Printf("ERROR: create wav encoder failed: filename=%v, error=%v\n", wavFileName, err)
				return err
			}

			session.SetAttr(keyRawFile, rawFile)
			session.SetAttr(keyWavFile, wavFile)
			session.SetAttr(keyRawFileName, rawFileName)
			session.SetAttr(keyWavEncoder, wavEncoder)
		}

		rawFile, _ = session.GetAttr(keyRawFile).(*os.File)
		_, err := rawFile.Write(pkt.Data)
		if err != nil {
			log.Printf("ERROR: write data to raw file: callid=%v, filename=%v, error=%v\n",
				callId, rawFileName, err,
			)
			return err
		}

		_, err = wavEncoder.Write(pkt.Data)
		if err != nil {
			log.Printf("ERROR: write data to wav file: callid=%v, filename=%v, error=%v\n",
				callId, wavFileName, err,
			)
			return err
		}
		log.Printf("go data %v bytes, callid=%v\n", pkt.Len, callId)
		return nil
	default:
		log.Printf("ERROR: invalid packet type: %v\n", pkt.Type)
		return ErrInvalidPacket
	}
	return nil
}

func main() {
	srvConf := knet.NewTCPServerConfig()
	//srvConf.MaxConnection = 2

	srv := knet.NewTCPServer(context.Background(), srvConf)
	srv.SetProtocol(&AudioProtocol{})
	srv.SetIoHandler(&srvAudioHandler{})

	addr := "0.0.0.0:9000"
	go func() {
		http.ListenAndServe(addr, nil)
	}()

	exeFilePath, err := os.Executable()
	if err != nil {
		log.Printf("ERROR: unable to get exe path: %v\n", err)
		os.Exit(1)
	}

	homeDir = path.Dir(exeFilePath)
	dataDir = path.Join(homeDir, "audios")

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.Mkdir(dataDir, os.ModePerm); err != nil {
			log.Printf("ERROR: unable to create dir: %v, error=%v\n", dataDir, err)
			os.Exit(2)
		}
	}

	log.Printf("server started, addr=%v\n", addr)
	srv.ListenAndServe(addr)
}
