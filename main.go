package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type Server struct {
	clients map[net.Conn]bool
	mu      sync.Mutex
}

func main() {
	listener, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatal("Error Listening....", err)
	}
	defer listener.Close()

	server := &Server{
		clients: make(map[net.Conn]bool),
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn", err)
			continue
		}

		server.mu.Lock()
		server.clients[conn] = true
		server.mu.Unlock()

		fmt.Print("v", server.clients[conn])

		go server.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {

			if err == io.EOF {
				log.Println("Client disconnected!!!")
			} else {
				log.Printf("Read error: %v", err)
			}
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			return
		}

		ackMsg := strings.TrimSpace(message)
		response := fmt.Sprintf("ACK: %s\n", ackMsg)

		_, err = conn.Write([]byte(response))
		if err != nil {
			log.Printf("Server write error: %v", err)
			return
		}
	}
}
