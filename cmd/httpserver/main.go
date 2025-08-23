package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const PORT = 42069

func main() {
	server, err := server.Serve(PORT, func(w io.Writer, req *request.Request) *server.HandlerError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandlerError{
				StatusCode: response.BadRequest,
				Message:    "Your problem is not my problem\n",
			}
		case "/myproblem":
			return &server.HandlerError{
				StatusCode: response.InternalError,
				Message:    "Woopsie, my bad\n",
			}
		default:
			w.Write([]byte("All good, frfr\n"))
			return nil
		}
	})
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	log.Println("Server is running on port", PORT)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
