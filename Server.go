package gn

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type ConnectHandler func(conn net.Conn)
type DisconnectHandler func(conn net.Conn)
type PacketHandler func(conn net.Conn, p Packet)
type ErrorHandler func(conn net.Conn, err error)
type CloseHandler func(conn net.Conn)

type Server struct {
	connections       []net.Conn
	connectHandler    *ConnectHandler
	disconnectHandler *DisconnectHandler
	packetHandler     *PacketHandler
	errorHandler      *ErrorHandler
	closeHandler      *CloseHandler
}

func (s Server) Listen(port string) {
	// init http server on provided port
	http.ListenAndServe(":"+port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// upgrade http request to websocket
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			// handle connection error
			fmt.Fprintln(os.Stderr, err)
		}
		// add new connection to list
		s.connections = append(s.connections, conn)
		// call connect handler
		if s.connectHandler != nil {
			(*s.connectHandler)(conn)
		}
		//fmt.Println("Client Connected")

		// start goroutine to handle connection
		go func() {
			defer conn.Close() // close connection on goroutine exit

			// loop and read data until an error is encountered
			for {
				msg, _, err := wsutil.ReadClientData(conn)
				if err != nil {
					// handle read error
					fmt.Fprintln(os.Stderr, err)
					break
				}
				// parse message
				packet := Load(msg)
				if s.packetHandler != nil {
					var handler = *(s.packetHandler)
					handler(conn, *packet)
				}
			}
		}()
	}))
}

func (s Server) Write(conn net.Conn, p Packet) {
	err := wsutil.WriteServerMessage(conn, ws.OpBinary, p.Build())
	if err != nil {
		// handle write error
		fmt.Fprintln(os.Stderr, err)
	}
}

func (s Server) Broadcast(p Packet) {
	// loop through all connections and send message
	for _, conn := range s.connections {
		err := wsutil.WriteServerMessage(conn, ws.OpBinary, p.Build())
		if err != nil {
			// handle write error
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func (s *Server) OnConnect(handler ConnectHandler) {
	s.connectHandler = &handler
}

func (s *Server) OnDisconnect(handler DisconnectHandler) {
	s.disconnectHandler = &handler
}

func (s *Server) OnData(handler PacketHandler) {
	s.packetHandler = &handler
}

func (s *Server) OnError(handler ErrorHandler) {
	s.errorHandler = &handler
}

func (s *Server) OnClose(handler CloseHandler) {
	s.closeHandler = &handler
}
