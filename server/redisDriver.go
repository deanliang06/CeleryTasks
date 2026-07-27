package main

import (
	"io"
	"net"
	"slices"
	"strconv"
)

type redisConn struct {
	buffer    []byte
	conn      net.Conn
	redisHost string
}

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

func initConn(hostName string) (redisConn, error) {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		return redisConn{buffer: make([]byte, 0)}, err
	}

	return redisConn{conn: conn, redisHost: hostName}, nil
}

func (redisConn *redisConn) pushQueue(data []byte) error {
	payload := createQueuePayload("push", data)
	if _, err := redisConn.conn.Write(payload); err != nil {
		return err
	}
	return nil
}

func (redisConn *redisConn) pollQueue() ([]byte, error) {
	payload := createQueuePayload("poll", make([]byte, 0))
	if _, err := redisConn.conn.Write(payload); err != nil {
		return nil, err
	}
	readFrom := redisConn.readFromConnection()
	return readFrom, nil
}

func (redisConn *redisConn) addMap(id string, data []byte) error {
	payload := createMapPayload("addM", id, data)
	if _, err := redisConn.conn.Write(payload); err != nil {
		return nil
	}
	return nil
}

func (redisConn *redisConn) getMap(id string) ([]byte, error) {
	payload := createMapPayload("getM", id, make([]byte, 0))
	if _, err := redisConn.conn.Write(payload); err != nil {
		return nil, err
	}
	readFrom := redisConn.readFromConnection()
	return readFrom, nil
}

func createMapPayload(opType, id string, data []byte) []byte {
	mapSpecific := "*" + strconv.Itoa(len([]byte(id))) + "\r"
	dataSize := len(data) + len([]byte(id)) + len([]byte(mapSpecific))
	toConvert := opType + "*" + strconv.Itoa(dataSize) + "\n" + mapSpecific
	return slices.Concat([]byte(toConvert), []byte(id), data)
}

func createQueuePayload(opType string, data []byte) []byte {
	dataSize := len(data)
	toConvert := opType + "*" + strconv.Itoa(dataSize) + "\n"
	return slices.Concat([]byte(toConvert), data)
}

func (redisConn *redisConn) readFromConnection() []byte {
	total := redisConn.buffer
	bytes := make([]byte, 1000)

	parseSize := false
	var dataLength, startOfData int
	for {
		n, err := redisConn.conn.Read(bytes)
		total = append(total, bytes[:n]...)
		doneWithCommand := false
		for len(total) > 0 && startOfData >= 0 {
			if !parseSize {
				dataLength, startOfData = parseLength(total)
				parseSize = true
			}
			if parseSize && len(total)-startOfData >= dataLength {
				bytes = total[startOfData : startOfData+dataLength]
				redisConn.buffer = total[startOfData+dataLength:]
				parseSize = false
				doneWithCommand = true
				break
			}
		}

		if doneWithCommand {
			break
		}

		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}
	return bytes
}

func (redisConn *redisConn) Close() {
	redisConn.conn.Close()
}
