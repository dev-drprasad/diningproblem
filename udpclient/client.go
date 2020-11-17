package client

import (
	"net"
)

type Client struct {
	conn *net.UDPConn
}

func (c *Client) Init(address string) error {
	if c.conn == nil {
		s, err := net.ResolveUDPAddr("udp4", address)
		if err != nil {
			return err
		}
		conn, err := net.DialUDP("udp4", nil, s)
		if err != nil {
			return err
		}
		c.conn = conn
	}
	return nil
}

func (c *Client) Send(data []byte) error {
	_, err := c.conn.Write(append(data, []byte("\n")...))
	if err != nil {
		return err
	}
	return nil
}
