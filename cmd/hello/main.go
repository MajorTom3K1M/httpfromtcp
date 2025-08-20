package main

import (
	"fmt"
	"httpfromtcp/internal/request"
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

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

func main() {
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*",
		numBytesPerRead: 3,
	}
	r, err := request.RequestFromReader(reader)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Parsed Request: %+v\n", r)

}

func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos > len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	if n > cr.numBytesPerRead {
		n = cr.numBytesPerRead
		cr.pos -= n - cr.numBytesPerRead
	}
	return n, nil
}

func parseRequestLine(request string) (*RequestLine, error) {
	i := strings.Index(request, "\r\n")
	requestLine := request[:i]
	if i == -1 {
		return nil, fmt.Errorf("invalid request line: %s", request)
	}

	fields := strings.Fields(requestLine)
	if len(fields) != 3 {
		return nil, fmt.Errorf("invalid request line: %s", requestLine)
	}

	method, requestTarget, httpVersion := fields[0], fields[1], fields[2]

	if isUppercase(method) {
		return nil, fmt.Errorf("invalid method: %s", method)
	}

	cleanedVersion, err := parseHTTPVersion(httpVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP version: %s", httpVersion)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   cleanedVersion,
	}, nil
}

func parseHTTPVersion(version string) (string, error) {
	const pfx = "HTTP/"
	if !strings.HasPrefix(version, pfx) {
		return "", fmt.Errorf("invalid HTTP version: %s", version)
	}

	v := version[len(pfx):]
	if len(v) == 0 {
		return "", fmt.Errorf("invalid HTTP version: %s", version)
	}

	return v, nil
}

func isUppercase(str string) bool {
	if len(str) == 0 {
		return false
	}

	for _, r := range str {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
