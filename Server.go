package gn

import (
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type ReadyHandler func(Server)
type ConnectHandler func(Connection)
type DisconnectHandler func(Connection)
type PacketHandler func(Connection, Packet)
type ErrorHandler func(ServerError)
type CloseHandler func(Server)

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

func (s *Server) Listen(port string) error {
	var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// upgrade http request to websocket
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			if s.errorHandler != nil {
				(*s.errorHandler)(ServerError{err, nil, s})
			}
			return
		}
		// add new connection to list
		c := *NewConnection(conn, *s)
		s.connections = append(s.connections, c)
		// call connect handler
		if s.connectHandler != nil {
			(*s.connectHandler)(c)
		}

		// start goroutine to handle connection
		// TODO: communicate with these routines via channels
		go func() {
			defer c.Close() // close connection on goroutine exit

			// loop and read data until an error is encountered
			for {
				msg, _, readErr := wsutil.ReadClientData(conn)
				if readErr != nil {
					// handle read error
					if s.errorHandler != nil {
						(*s.errorHandler)(ServerError{readErr, &c, s})
					}
					break
				}
				// parse message
				packet, parseErr := Load(msg)
				if parseErr != nil {
					// handle parse error
					if s.errorHandler != nil {
						(*s.errorHandler)(ServerError{parseErr, &c, s})
					}
					break
				}
				// call packet handler
				if s.packetHandler != nil {
					(*s.packetHandler)(c, *packet)
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
	err := serv.ListenAndServe()
	if err == nil {
		s.port = p
		s.serv = serv
		// call ready handler
		if s.readyHandler != nil {
			(*s.readyHandler)(*s)
		}
	}
	return err
}

func (s Server) Broadcast(p Packet) error {
	// loop through all connections and send message
	for _, c := range s.connections {
		err := c.Write(p)
		if err != nil {
			return err
		}
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
}

func (s *Server) OnReady(handler ReadyHandler) {
	s.readyHandler = &handler
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
