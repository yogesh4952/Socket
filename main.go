package main

import (
	"bufio"
	"io"
	"log"
	"net"
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
		go server.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	defer func() {
		s.mu.Lock()
		s.clients[conn] = false
		delete(s.clients, conn)
		s.mu.Unlock()
	}()
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {

			if err == io.EOF {
				log.Println("Client disconnected!!!")
			} else {
				log.Printf("Read error: %v", err)
			}

			return
		}

		s.broadcast(conn, message)

	}
}

func (s *Server) broadcast(sender net.Conn, message string) {
	s.mu.Lock()
	targets := make([]net.Conn, 0, len(s.clients))

	for conn := range s.clients {
		if conn != sender {
			targets = append(targets, conn)
		}
	}
	s.mu.Unlock()

	for _, conn := range targets {
		if _, err := conn.Write([]byte(message)); err != nil {
			log.Printf("write to %v failed: %v", conn.RemoteAddr(), err)
		}
	}
}
