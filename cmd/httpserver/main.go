package main

import (
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const PORT = 42069

func main() {
	server, err := server.Serve(PORT)
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
