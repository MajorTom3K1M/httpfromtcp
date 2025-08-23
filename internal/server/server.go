package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type Server struct {
	Listener net.Listener
	Handler  Handler
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

func Serve(port int, handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		Handler:  handler,
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

			go s.handle(c)
		}(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		WriteErrorResponse(conn, &HandlerError{
			StatusCode: response.BadRequest,
			Message:    fmt.Sprintf("error reading request: %v", err),
		})
		return
	}

	writter := response.NewResponseWriter(conn)
	s.Handler(writter, req)
}
