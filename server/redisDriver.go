package main

import (
	"io"
	"net"
	"slices"
	"strconv"
	"strings"
)

func parseLength(data []byte) (int, int) {
	stringified := string(data)
	var start int
	for i, char := range stringified {
		switch {
		case char == '*':
			start = i
		case char == '\n':
			len, err := strconv.Atoi(string(data[start+1 : i]))
			if err != nil {
				panic(err)
			}
			return len, i + 1
		}
	}
	return -1, -1
}

type redisConn struct {
	conn      net.Conn
	redisHost string
}

func initConn(hostName string) (redisConn, error) {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		return redisConn{}, err
	}

	return redisConn{conn: conn, redisHost: hostName}, nil
}

func (redisConn *redisConn) pushQueue(data []byte) error {
	payload := createPayload("push", data)
	if _, err := redisConn.conn.Write(payload); err != nil {
		return err
	}
	return nil
}

func (redisConn *redisConn) pollQueue() ([]byte, error) {
	payload := createPayload("poll", make([]byte, 0))
	if _, err := redisConn.conn.Write(payload); err != nil {
		return nil, err
	}
	readFrom := redisConn.readFromConnection()
	return readFrom, nil
}

func createPayload(opType string, data []byte) []byte {
	dataSize := len(data)
	toConvert := opType + "*" + strconv.Itoa(dataSize) + "\n"
	return slices.Concat([]byte(toConvert), data)
}
func (redisConn *redisConn) readFromConnection() []byte {
	total := make([]byte, 0)
	bytes := make([]byte, 1000)

	parseSize := false
	var dataLength, startOfData int
	for {
		n, err := redisConn.conn.Read(bytes)
		total = append(total, bytes[:n]...)
		for len(total) > 0 && startOfData != -1 {
			if strings.Contains(string(total), "\n") && !parseSize {
				dataLength, startOfData = parseLength(total)
				parseSize = true
			}
			if parseSize && len(total)-startOfData >= dataLength {
				total = total[startOfData+dataLength:]
				bytes = make([]byte, 1000)
				parseSize = false
			}
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
