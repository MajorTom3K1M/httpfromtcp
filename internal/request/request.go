package request

import (
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	_, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: RequestLine{
			HttpVersion:   "HTTP/1.1",
			RequestTarget: "/",
			Method:        "GET",
		},
	}, nil
}

func parseRequestLine(requestLine string) (*RequestLine, error) {
	parts := strings.Split(requestLine, "\r\n")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line: %s", requestLine)
	}

	return &RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   parts[2],
	}, nil
}
