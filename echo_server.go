package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"strings"
)

type EchoServer struct {
	RelayAddress string
}

func main() {
	relayAddress := flag.String("relay", "localhost:8080", "Relay server address")
	flag.Parse()

	echoserver := EchoServer{RelayAddress: *relayAddress}
	echoserver.ConnectToRelay()
}

func (this EchoServer) ConnectToRelay() {
	fmt.Println("Echo server connecting with Relay server at :", this.RelayAddress)

	conn, err := net.Dial("tcp", this.RelayAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	if _, err := conn.Write(append([]byte("new-server"), '\n')); err != nil {
		fmt.Println("Err writing client connect", err)
		return
	}

	for {
		reader := bufio.NewReader(conn)
		data, err := reader.ReadString('\n')
		if err != nil { // EOF, or worse
			fmt.Println("error is", err)
			return
		}
		data = strings.Trim(data, "\n")
		go this.ResponseFromRelay(data)
	}
}

func (this EchoServer) ResponseFromRelay(resp string) {
	fmt.Println("Handling response from relay server")

	switch {

	case strings.Contains(resp, "connected"):
		result := strings.Split(resp, "@")
		fmt.Println("estabilished relay address: ",result[1])

	case strings.Contains(resp, "client"):
		//Dial to relay server
		conn, err := net.Dial("tcp", this.RelayAddress)
		if err != nil {
			fmt.Println("Err dialing to Relay server", err)
			return
		}

		if _, err := conn.Write(append([]byte(resp), '\n')); err != nil {
			fmt.Println("Err writing to client connection", err)
			return
		}
		//Read the data from conn and writes it back to conn
		io.Copy(conn, conn)

	default:
		fmt.Println("Error from Relay server:", resp)
		return

	}
}
