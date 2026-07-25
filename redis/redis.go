package main

import (
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	fmt.Println(conn.LocalAddr())
	bytes := make([]byte, 80000)
	conn.Read(bytes)
	fmt.Println(string(bytes))
	conn.Close()
}

func main() {
	ln, err := net.Listen("tcp", ":6379")
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
