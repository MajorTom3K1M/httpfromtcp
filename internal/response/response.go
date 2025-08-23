package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	OK            StatusCode = 200
	BadRequest    StatusCode = 400
	Unauthorized  StatusCode = 401
	Forbidden     StatusCode = 403
	NotFound      StatusCode = 404
	InternalError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var line string
	switch statusCode {
	case OK:
		line = "HTTP/1.1 200 OK\r\n"
	case BadRequest:
		line = "HTTP/1.1 400 Bad Request\r\n"
	case NotFound:
		line = "HTTP/1.1 404 Not Found\r\n"
	case InternalError:
		line = "HTTP/1.1 500 Internal Server Error\r\n"
	default:
		return fmt.Errorf("unsupported status code: %d", statusCode)
	}

	_, err := io.WriteString(w, line)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	defaultHeaders := headers.NewHeaders()
	defaultHeaders.Set("Content-Length", strconv.Itoa(contentLen))
	defaultHeaders.Set("Connection", "Close")
	defaultHeaders.Set("Content-Type", "text/plain")

	return defaultHeaders
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	if len(headers) == 0 {
		return fmt.Errorf("no headers to write")
	}

	var headerLines string
	for key, value := range headers {
		headerLines += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	headerLines += "\r\n"

	_, err := io.WriteString(w, headerLines)

	return err
}
