package server

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func WriteErrorResponse(w io.Writer, err *HandlerError) {
	if err == nil {
		return
	}

	if writeErr := response.WriteStatusLine(w, err.StatusCode); writeErr != nil {
		log.Fatalf("status line write error: %v", writeErr)
		response.WriteStatusLine(w, response.InternalError)
		return
	}

	headers := response.GetDefaultHeaders(len(err.Message))
	if writeErr := response.WriteHeaders(w, headers); writeErr != nil {
		log.Fatalf("headers write error: %v", writeErr)
		response.WriteStatusLine(w, response.InternalError)
		return
	}

	io.WriteString(w, err.Message)
}
