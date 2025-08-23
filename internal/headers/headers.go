package headers

import (
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	i := strings.Index(string(data), "\r\n")
	if i == -1 {
		return 0, false, nil
	}

	headersLine := string(data[:i])
	n = i + 2

	if headersLine == "" || i == 0 {
		return n, true, nil
	}

	headersLine = strings.TrimSpace(headersLine)

	k, v, ok := strings.Cut(headersLine, ":")
	if !ok {
		return 0, false, fmt.Errorf("invalid header format: %s", headersLine)
	}

	if strings.Contains(k, " ") {
		return 0, false, fmt.Errorf("invalid header key spacing: %s", k)
	}

	v = strings.TrimSpace(v)

	regexpKey := regexp.MustCompile(`^[a-zA-Z0-9!#$%&'*+.^_` + "`" + `|~-]+$`)
	if !regexpKey.MatchString(k) {
		return 0, false, fmt.Errorf("invalid header key: %s", k)
	}

	key := strings.ToLower(k)
	if prev, ok := h[key]; ok && prev != "" {
		v = prev + ", " + v
	}
	h[key] = v

	return n, false, nil
}

func (h Headers) Set(key string, value string) {
	key = strings.ToLower(key)
	v, ok := h[key]
	if ok {
		value = v + ", " + value
	}
	h[key] = value
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Override(key, value string) {
	h[strings.ToLower(key)] = value
}

func (h Headers) Del(key string) {
	delete(h, strings.ToLower(key))
}

func NewHeaders() Headers {
	return make(Headers)
}
