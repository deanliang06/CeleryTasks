package main

import (
	"errors"
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
func handleConnection(conn net.Conn) {
	total := make([]byte, 0)
	bytes := make([]byte, 1000)

	parseSize := false
	var dataLength, startOfData int
	for {
		n, err := conn.Read(bytes)
		if n > 0 {
			total = append(total, bytes[:n]...)
			if parseSize && len(total)-startOfData >= dataLength {
				output := handleRead(bytes, startOfData)
				if output != nil {
					conn.Write(formatOutput(output))
				}

				total = make([]byte, 0)
				bytes = make([]byte, 1000)
				parseSize = false
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
}

func formatOutput(data []byte) []byte {
	dataLength := len(data)
	preFix := "*" + strconv.Itoa(dataLength) + "\n"
	return slices.Concat([]byte(preFix), data)
}

func handleRead(bytes []byte, start int) []byte {
	stringMsg := string(bytes)
	fmt.Println(stringMsg)
	typeMsg := stringMsg[:4]
	data := stringMsg[start:]
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
