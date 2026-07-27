package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"strconv"
	"strings"
)

func parseLength(data []byte) (int, int, error) {
	stringified := string(data)
	var start int
	for i, char := range stringified {
		switch {
		case char == '*':
			start = i
		case char == '\n':
			return i - start, i + 1, nil
		}
	}
	return 0, 0, errors.New("WTF is going on")
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
	fmt.Println(dataSize)
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
		if n > 0 {
			total = append(total, bytes[:n]...)
			if parseSize && len(total)-startOfData >= dataLength {
				break
			}
			if strings.Contains(string(total), "\n") && !parseSize {
				dataLength, startOfData, err = parseLength(total)
				if err != nil {
					panic(err)
				}
				parseSize = true
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
