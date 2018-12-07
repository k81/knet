package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/k81/knet"
	"github.com/k81/knet/examples/echo"
)

type srvEchoHandler struct {
	knet.IoHandlerAdapter
}

func (h *srvEchoHandler) OnConnected(session *knet.IoSession) error {
	log.Printf("session connected, remote_addr=%v", session.RemoteAddr())
	//m := echo.NewEchoMessage("welcome to the echo server, enjoy it!")
	//session.Send(m)
	return nil
}

func (h *srvEchoHandler) OnDisconnected(session *knet.IoSession) {
	log.Printf("session disconnected, remote_addr=%v, stats=%v", session.RemoteAddr(), session.String())
}

func (h *srvEchoHandler) OnError(session *knet.IoSession, err error) {
	log.Printf("sesson error: %v", err)
}

func (h *srvEchoHandler) OnMessage(session *knet.IoSession, m knet.Message) error {
	echoMsg := m.(*echo.EchoMessage)
	log.Printf("RECV: msg=%v", echoMsg.Content)
	session.Send(mctx, m)
	return nil
}

func main() {
	srvConf := knet.NewTCPServerConfig()
	//srvConf.MaxConnection = 2

	srv := knet.NewTCPServer(mctx, srvConf)
	srv.SetProtocol(&echo.EchoProtocol{})
	srv.SetIoHandler(&srvEchoHandler{})

	go func() {
		http.ListenAndServe("127.0.0.1:8890", nil)
	}()

	addr := "127.0.0.1:8888"
	log.Printf("server started, addr=%v", addr)
	srv.ListenAndServe(addr)
}
