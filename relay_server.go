package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ServerConn struct {
	conn net.Conn
}

type RelayServer struct {
	lastPort           int
	serverConnMap      map[string]*ServerConn // key is server address
	clientConnMap      map[string]net.Conn    // key is client id randomly generated when it connects
	serverConnMapMutex sync.Mutex
	clientConnMapMutex sync.Mutex
}

func NewRelayServer() *RelayServer {
	return &RelayServer{
		serverConnMap:      make(map[string]*ServerConn),
		clientConnMap:      make(map[string]net.Conn),
		serverConnMapMutex: sync.Mutex{},
		clientConnMapMutex: sync.Mutex{},
	}
}

func main() {
	port := flag.Int("port", 8080, "Port relay server will listen on")
	flag.Parse()

	fmt.Println(fmt.Sprintf("Listening on port %d", *port))
	tcpRelay := NewRelayServer()
	tcpRelay.RelayServerListen(*port)
}

func (this *RelayServer) RelayServerListen(port int) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go this.handleRelayRequest(conn)
	}
}

func (this *RelayServer) handleRelayRequest(conn net.Conn) {
	reader := bufio.NewReader(conn)
	data, err := reader.ReadString('\n')
	if err != nil { // EOF, or worse
		fmt.Println("error is", err)
		return
	}
	fmt.Println("data read from connection is", data)
	data = strings.Trim(data, "\n")

	switch {
	case data == "new-server":
		this.newPublicRelayServer(conn)
	case strings.Contains(data, "client"):
		this.newClient(conn, data)
	default:
		if _, err := conn.Write(append([]byte("wrong request type"), '\n')); err != nil {
			fmt.Println("Err writing client connect", err)
			return
		}
	}
}

func (this *RelayServer) newPublicRelayServer(conn net.Conn) {
	//fmt.Println("Iam inside new relay fucn")
	port, err := this.FindPort()
	if err != nil {
		if _, err := conn.Write(append([]byte("Error creating relay"), '\n')); err != nil {
			fmt.Println("Err writing client connect", err)
			return
		}
		return
	}

	address := fmt.Sprintf("%s:%d", "localhost", port)
	serverConn := &ServerConn{conn: conn}
        //lock the server conn map when we write to it, so that no one reads from it then
	this.serverConnMapMutex.Lock()
	this.serverConnMap[address] = serverConn
	this.serverConnMapMutex.Unlock()

	go this.PublicRelayServerlisten(address)

	msg := "connected" + "@" + address
	if _, err := conn.Write(append([]byte(msg), '\n')); err != nil {
		fmt.Println("Err writing client connect", err)
		return
	}
}

func (this *RelayServer) newClient(conn net.Conn, id string) {
	//lock the map while reading and unlock afetr reading is done
	this.clientConnMapMutex.Lock()
	clientConn, ok := this.clientConnMap[id]
	this.clientConnMapMutex.Unlock()
	if !ok {
		fmt.Println("Error reading from map")
		if _, err := conn.Write(append([]byte("error relaying connection"), '\n')); err != nil {
			fmt.Println("Err writing client connect", err)
			return
		}
		return
	}

	relayConn(clientConn, conn)
}

func (this *RelayServer) FindPort() (int, error) {
	fmt.Println("Iam inside available port")
	if this.lastPort == 0 {
		this.lastPort = 5000
	} else {
		this.lastPort++
	}
	if this.lastPort > 7000 {
		return 0, errors.New("No ports avaliable")
	}
	return this.lastPort, nil
}

func (this *RelayServer) PublicRelayServerlisten(address string) {

	l, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		fmt.Println("New connection on public relay server", address)
		if err != nil {
			panic(err)
		}

		go this.handlePublicRelayRequest(address, conn)
	}
}

func (this *RelayServer) handlePublicRelayRequest(address string, conn net.Conn) {
	//fmt.Println("Iam inside notify server")
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	var id string
	//Generating random client id when it connects with the relay server
	id = "client" + strconv.Itoa(r1.Intn(1000))

	//lock the client conn map when we write to it
	this.clientConnMapMutex.Lock()
	this.clientConnMap[id] = conn
	this.clientConnMapMutex.Unlock()

        //Lock the serverconnmap while reading from it
	this.serverConnMapMutex.Lock()
	serverConn, ok := this.serverConnMap[address]
	this.serverConnMapMutex.Unlock()
	if !ok {
		fmt.Println("The server not found in map")
		// server not present, close client conn
		conn.Close()
		return
	}

	if _, err := serverConn.conn.Write(append([]byte(id), '\n')); err != nil {
		fmt.Println("Err writing client connect", err)
		return
	}

	fmt.Println("Notified server of new conn", address)
}

func relayConn(clientConn net.Conn, serverConn net.Conn) {
	//Copies from client to server and vice versa
	go io.Copy(clientConn, serverConn)
	go io.Copy(serverConn, clientConn)
}
