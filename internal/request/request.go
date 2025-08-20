package request

import (
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	State       RequestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type RequestState string

const (
	InitialState RequestState = "Initial"
	HeadersState RequestState = "Headers"
	DoneState    RequestState = "Done"
)

const bufferSize = 1024

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, bufferSize)
	readToIndex := 0

	for !request.isDone() {
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)

			copy(newBuf, buf[:readToIndex])

			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.State != DoneState {
					return nil, fmt.Errorf("incomplete request")
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		if numBytesParsed > 0 {
			copy(buf, buf[numBytesParsed:readToIndex])
			readToIndex -= numBytesParsed
		}

		if numBytesRead == 0 && numBytesParsed == 0 {
			return nil, io.ErrNoProgress
		}
	}

	return request, nil
}

func parseRequestLine(request string) (*RequestLine, int, error) {
	i := strings.Index(request, "\r\n")
	if i == -1 {
		return nil, 0, nil
	}

	requestLine := request[:i]

	fields := strings.Fields(requestLine)
	if len(fields) != 3 {
		return nil, 0, fmt.Errorf("invalid request line: %s", requestLine)
	}

	method, requestTarget, httpVersion := fields[0], fields[1], fields[2]

	if !isUppercase(method) {
		return nil, 0, fmt.Errorf("invalid method: %s", method)
	}

	cleanedVersion, err := parseHTTPVersion(httpVersion)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid HTTP version: %s", httpVersion)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   cleanedVersion,
	}, len(request[:i]) + len("\r\n"), nil
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

	parts := strings.Split(v, ".")
	if len(parts) > 2 {
		return "", fmt.Errorf("invalid HTTP version: %s", version)
	}

	if _, err := strconv.Atoi(parts[0]); err != nil {
		return "", fmt.Errorf("invalid HTTP version: %s", version)
	}

	if len(parts) == 2 {
		if _, err := strconv.Atoi(parts[1]); err != nil {
			return "", fmt.Errorf("invalid HTTP version: %s", version)
		}
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

func (r *Request) parse(data []byte) (int, error) {
	read := 0

	for {
		switch r.State {
		case InitialState:
			requestLine, n, err := parseRequestLine(string(data[read:]))
			if err != nil {
				return 0, err
			}

			if n == 0 {
				return read, nil
			}

			r.RequestLine = *requestLine
			r.State = HeadersState

			read += n
		case HeadersState:
			n, done, err := r.Headers.Parse(data[read:])
			if err != nil {
				return 0, err
			}

			if done {
				read += n
				r.State = DoneState
				return read, nil
			}

			if n == 0 {
				return read, nil
			}

			read += n
		case DoneState:
			return read, fmt.Errorf("error: trying to read data in a done state")
		}
	}
}

func (r *Request) isDone() bool {
	return r.State == DoneState
}

func newRequest() *Request {
	return &Request{
		Headers: headers.NewHeaders(),
		State:   InitialState,
	}
}
