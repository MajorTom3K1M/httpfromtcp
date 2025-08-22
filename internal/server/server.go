package server

import (
	"bufio"
	"fmt"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type Server struct {
	Listener net.Listener
	State    ServerState
	Port     int
	Closed   atomic.Bool
	wg       sync.WaitGroup
}

type ServerState string

const (
	ServerStateRunning ServerState = "running"
	ServerStateClosing ServerState = "closing"
	ServerStateStopped ServerState = "stopped"
	ServerStateError   ServerState = "error"
)

func Serve(port int) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		Port:     port,
		Listener: ln,
		State:    ServerStateRunning,
	}
	go s.listen()

	return s, nil
}

func (s *Server) Close() error {
	if s == nil || s.Listener == nil {
		return fmt.Errorf("server is not initialized or already closed")
	}

	if !s.Closed.CompareAndSwap(false, true) {
		return nil
	}

	s.State = ServerStateClosing

	err := s.Listener.Close()

	s.wg.Wait()

	if err != nil {
		s.State = ServerStateError
		return fmt.Errorf("failed to close listener: %w", err)
	}
	s.State = ServerStateStopped
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()

		if err != nil {
			if s.Closed.Load() {
				return
			}

			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Printf("temporary error accepting connection: %v", err)
				continue
			}
			log.Printf("accept error (stopping): %v", err)
			s.State = ServerStateError
			return
		}

		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()
			s.handle(c)
		}(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	bw := bufio.NewWriter(conn)
	if err := response.WriteStatusLine(bw, response.OK); err != nil {
		log.Printf("error writing status line: %v", err)
		return
	}

	hdrs := response.GetDefaultHeaders(0)
	if err := response.WriteHeaders(bw, hdrs); err != nil {
		log.Printf("error writing headers: %v", err)
		return
	}

	if err := bw.Flush(); err != nil {
		log.Printf("error flushing response: %v", err)
		return
	}

}
