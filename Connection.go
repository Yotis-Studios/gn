package gn

import (
	"net"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Connection struct {
	conn net.Conn
	serv Server
}

func NewConnection(conn net.Conn, serv Server) *Connection {
	c := new(Connection)
	c.conn = conn
	c.serv = serv
	return c
}

func (c Connection) Write(p Packet) error {
	msg, err := p.Build()
	if err != nil {
		return err
	}
	err = wsutil.WriteServerMessage(c.conn, ws.OpBinary, msg)
	return err
}

func (c Connection) Close() error {
	// remove from server list
	conns := c.serv.connections
	for i, conn := range conns {
		if c.conn == conn.conn {
			n := len(conns) - 1
			conns[i] = conns[n]
			c.serv.connections = conns[:n]
		}
	}
	return c.conn.Close()
}
