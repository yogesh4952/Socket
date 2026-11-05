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

type client struct {
	name string
	conn net.Conn
}

type Server struct {
	clients map[net.Conn]*client
	mu      sync.Mutex
}

func main() {
	listener, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatal("Error Listening....", err)
	}
	defer listener.Close()

	server := &Server{
		clients: make(map[net.Conn]*client),
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn", err)
			continue
		}
		go server.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	defer func() {
		s.mu.Lock()

		delete(s.clients, conn)
		s.mu.Unlock()
	}()
	reader := bufio.NewReader(conn)

	conn.Write([]byte("Enter username: "))
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading username: ", err)
	}
	name = strings.TrimSpace(name)

	if name == "" {
		name = conn.RemoteAddr().String()
	}

	s.mu.Lock()
	s.clients[conn] = &client{name: name, conn: conn}
	s.mu.Unlock()
	log.Printf("Registered %q from %s", name, conn.RemoteAddr())
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

		s.broadcast(conn, name, message)

	}
}

func (s *Server) broadcast(senderConn net.Conn, SenderName, message string) {
	s.mu.Lock()
	targets := make([]*client, 0, len(s.clients))

	for c, cl := range s.clients {
		if c != senderConn {
			targets = append(targets, cl)
		}
	}
	s.mu.Unlock()

	ack := fmt.Sprintf("%s: %s", SenderName, message)
	for _, cl := range targets {
		if _, err := cl.conn.Write([]byte(ack)); err != nil {
			log.Printf("write to %v failed: %v", cl, err)
		}
	}
}
