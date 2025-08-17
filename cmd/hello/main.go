package main

import (
	"fmt"
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

func main() {
	parse, err := parseRequestLine("GET / HTTP/1.1\r\nHost: example.com\r\nUser-Agent: test\r\n\r\n")
	if err != nil {
		fmt.Println("Error parsing request line:", err)
		return
	}
	fmt.Printf("Parsed Request: %+v\n", parse)
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
