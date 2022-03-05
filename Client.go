package gn

import (
	"context"
	"net"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Client struct {
	conn net.Conn
}

func (c *Client) Connect(addr string) error {
	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), addr)
	if err != nil {
		return err
	}
	c.conn = conn

	return nil
}

func (c Client) Disconnect() error {
	return c.conn.Close()
}

func (c Client) Send(p Packet) (err error) {
	var msg []byte = p.raw
	if msg == nil {
		msg, err = p.Build()
		if err != nil {
			return err
		}
	}
	err = wsutil.WriteServerMessage(c.conn, ws.OpBinary, msg)
	return err
}
