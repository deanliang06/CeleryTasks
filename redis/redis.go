package main

import (
	"fmt"
	"io"
	"net"
	"slices"
	"strconv"
	"strings"
	"sync"
)

type LinkNode struct {
	data []byte
	next *LinkNode
}

type Queue struct {
	head *LinkNode
	tail *LinkNode
}

var queue = Queue{head: nil, tail: nil}
var taskMap = Map{mut: sync.Mutex{}, actMap: make(map[string][]byte)}

type Map struct {
	mut    sync.Mutex
	actMap map[string][]byte
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
func handleConnection(conn net.Conn) {
	total := make([]byte, 0)
	bytes := make([]byte, 1000)

	parseSize := false
	var dataLength, startOfData int
	for {
		n, err := conn.Read(bytes)

		total = append(total, bytes[:n]...)
		for len(total) > 0 && startOfData != -1 {
			if strings.Contains(string(total), "\n") && !parseSize {
				dataLength, startOfData = parseLength(total)
				parseSize = true
			}

			if parseSize && len(total)-startOfData >= dataLength {
				output := handleRead(total, startOfData, dataLength)
				if output != nil {
					conn.Write(formatOutput(output))
				}

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
}

func formatOutput(data []byte) []byte {
	dataLength := len(data)
	preFix := "*" + strconv.Itoa(dataLength) + "\n"
	return slices.Concat([]byte(preFix), data)
}

func handleRead(bytes []byte, start, length int) []byte {
	stringMsg := string(bytes)
	typeMsg := stringMsg[:4]
	data := stringMsg[start : start+length]
	fmt.Println(typeMsg, data)
	return handleMsg(typeMsg, []byte(data))
}

func handleMsg(msgType string, data []byte) []byte {
	switch {
	case msgType == "poll":
		return pollQueue()
	case msgType == "push":
		return pushQueue(data)
	default:
		return nil
	}
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err.Error())
		}

		go handleConnection(conn)
	}
}
