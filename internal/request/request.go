package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	State       RequsetState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type RequsetState string

const (
	InitialState RequsetState = "Initial"
	DoneState    RequsetState = "Done"
)

const bufferSize = 8

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
			if err == io.EOF && readToIndex == 0 {
				request.State = DoneState
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
	}, len(request), nil
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
	if r.State == InitialState {
		requestLine, readN, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}

		if readN == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.State = DoneState

		return readN, nil
	} else if r.State == DoneState {
		return 0, fmt.Errorf("error: trying to read data in a done state")
	}

	return 0, fmt.Errorf("error: trying to read unsupported state")
}

func (r *Request) isDone() bool {
	return r.State == DoneState
}

func newRequest() *Request {
	return &Request{
		State: InitialState,
	}
}
