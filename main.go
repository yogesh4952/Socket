package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type client struct {
	name string
	conn net.Conn
}

type Server struct {
	byConn map[net.Conn]*client
	byName map[string]*client
	mu     sync.Mutex
}

func main() {
	listener, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatal("Error Listening....", err)
	}
	defer listener.Close()

	server := &Server{
		byConn: make(map[net.Conn]*client),
		byName: make(map[string]*client),
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
		s.unregister(conn)
	}()
	reader := bufio.NewReader(conn)

	conn.Write([]byte("Enter username: "))
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading username: ", err)
		return
	}
	name = strings.TrimSpace(name)

	if name == "" {
		name = conn.RemoteAddr().String()
	}
	cl := &client{name: name, conn: conn}

	if err := s.register(cl); err != nil {
		conn.Write([]byte(err.Error() + "\n"))
		log.Printf("Rejected %s: %v", conn.RemoteAddr(), err)
		return
	}
	log.Printf("Registered %q from %s", name, conn.RemoteAddr())
	for {
		message, err := reader.ReadString('\n')
		if err != nil {

			if err == io.EOF {
				log.Printf("%q disconnected!!!", cl.name)
			} else {
				log.Printf("Read error: %v", err)
			}

			return
		}

		s.broadcast(conn, message)

	}
}

func (s *Server) unregister(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cl, exists := s.byConn[conn]; exists {
		delete(s.byConn, cl.conn)
		delete(s.byName, cl.name)
	}
}

func (s *Server) register(cl *client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[cl.name]; exists {
		return fmt.Errorf("username %q already taken", cl.name)
	}
	s.byConn[cl.conn] = cl
	s.byName[cl.name] = cl
	return nil
}

func (s *Server) broadcast(senderConn net.Conn, message string) {
	s.mu.Lock()
	sender := s.byConn[senderConn]
	targets := make([]*client, 0, len(s.byConn))

	for c, cl := range s.byConn {
		if c != senderConn {
			targets = append(targets, cl)
		}
	}
	s.mu.Unlock()

	ack := fmt.Sprintf("%s: %s", sender.name, message)
	for _, cl := range targets {

		if err := cl.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
			log.Printf("write to %v failed:%v", cl, err)
		}

		if _, err := cl.conn.Write([]byte(ack)); err != nil {
			log.Printf("write to %v failed: %v", cl, err)
		}
	}
}
