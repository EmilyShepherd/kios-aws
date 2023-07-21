package main

import (
	"net"
)

type SystemSocket struct {
	conn net.Conn
}

const (
	SystemSocketPath = "/run/system.sock"

	CmdShutdown         = byte(1)
	CmdContinueShutdown = byte(2)
	CmdRestartKubelet   = byte(3)
	CmdRestartCrio      = byte(4)

	EventShutdown = byte(1)
)

func NewSystemSocket() (*SystemSocket, error) {
	conn, err := net.Dial("unix", SystemSocketPath)
	return &SystemSocket{
		conn: conn,
	}, err
}

func (s *SystemSocket) SendCmd(cmd byte) error {
	_, err := s.conn.Write([]byte{cmd})
	return err
}
