package gn

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type ReadyHandler func(serv Server)
type ConnectHandler func(conn Connection)
type DisconnectHandler func(conn Connection)
type PacketHandler func(conn Connection, p Packet)
type ErrorHandler func(conn Connection, err error)
type CloseHandler func(serv Server)

type Server struct {
	connections       []Connection
	serv              *http.Server
	port              string
	readyHandler      *ReadyHandler
	connectHandler    *ConnectHandler
	disconnectHandler *DisconnectHandler
	packetHandler     *PacketHandler
	errorHandler      *ErrorHandler
	closeHandler      *CloseHandler
}

func (s Server) Listen(port string) error {
	var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// upgrade http request to websocket
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			// TODO: handle connection error
			fmt.Fprintln(os.Stderr, err)
			return
		}
		// add new connection to list
		c := *NewConnection(conn, s)
		s.connections = append(s.connections, c)
		// call connect handler
		if s.connectHandler != nil {
			(*s.connectHandler)(c)
		}
		//fmt.Println("Client Connected")

		// start goroutine to handle connection
		// TODO: communicate with these routines via channels
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
				//fmt.Println("received: ", msg)
				// parse message
				packet := Load(msg)
				// call packet handler
				if s.packetHandler != nil {
					var handler = *(s.packetHandler)
					handler(c, *packet)
				}
			}
		}()
	})

	// init http server on provided port
	p := ":" + port
	serv := &http.Server{
		Addr:    p,
		Handler: handler,
	}
	err := s.serv.ListenAndServe()
	if err == nil {
		//fmt.Println("Server started on port: ", port)
		s.port = p
		s.serv = serv
		// call ready handler
		if s.readyHandler != nil {
			(*s.readyHandler)(s)
		}
	}
	return err
}

func (s Server) Broadcast(p Packet) error {
	// loop through all connections and send message
	for _, c := range s.connections {
		err := wsutil.WriteServerMessage(c.conn, ws.OpBinary, p.Build())
		return err
	}
	return nil // no error
}

func (s Server) Close() {
	// loop through all connections and close
	for _, c := range s.connections {
		c.conn.Close()
	}
	// close server
	s.serv.Close()
	// call close handler
	if s.closeHandler != nil {
		(*s.closeHandler)(s)
	}
	//fmt.Println("Server closed")
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
