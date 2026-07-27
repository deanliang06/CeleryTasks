package main

import (
	"io"
	"net"
	"slices"
)

type redisConn struct {
	conn      net.Conn
	redisHost string
}

var endCommand []byte = []byte("\r\r\r")

func initConn(hostName string) (redisConn, error) {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		return redisConn{}, err
	}

	return redisConn{conn: conn, redisHost: hostName}, nil
}

func (redisConn *redisConn) pushQueue(data []byte) error {
	typeString := "push\n"
	payload := slices.Concat([]byte(typeString), data)
	if _, err := redisConn.conn.Write(payload); err != nil {
		return err
	}
	return nil
}

func (redisConn *redisConn) pollQueue() ([]byte, error) {
	typeString := "poll\n"
	payload := []byte(typeString)
	if _, err := redisConn.conn.Write(payload); err != nil {
		return nil, err
	}
	readFrom := redisConn.readFromConnection()
	return readFrom, nil
}

func (redisConn *redisConn) readFromConnection() []byte {
	total := make([]byte, 0)
	bytes := make([]byte, 1000)

	for {
		n, err := redisConn.conn.Read(bytes)
		if n > 0 {
			total = append(total, bytes[:n]...)
		}
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}
	return total
}

func (redisConn *redisConn) Close() {
	redisConn.conn.Close()
}
