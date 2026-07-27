package main

import (
	"fmt"
	"io"
	"net"
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

func handleConnection(conn net.Conn) {
	fmt.Println(conn.LocalAddr())
	total := make([]byte, 0)
	bytes := make([]byte, 1000)

	defer conn.Close()
	for {
		n, err := conn.Read(bytes)
		if n > 0 {
			total = append(total, bytes[:n]...)
		}
		if err == io.EOF || n == 0 {
			break
		} else if err != nil {
			panic(err)
		}
	}

	output := handleRead(bytes)
	if output != nil {
		conn.Write(output)
	}
}

func handleRead(bytes []byte) []byte {
	stringMsg := strings.Split(string(bytes), "\n")
	fmt.Println(stringMsg)
	typeMsg := stringMsg[1]
	data := stringMsg[2]
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

		handleConnection(conn)
	}
}
