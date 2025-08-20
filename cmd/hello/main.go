package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"io"
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
