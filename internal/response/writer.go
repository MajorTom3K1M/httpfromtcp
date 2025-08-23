package response

import (
	"bufio"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type Resposne struct {
}

type Writer struct {
	StartLine string
	Headers   headers.Headers
	Body      []byte
	State     WriterState
	bw        *bufio.Writer
}

type WriterState string

const (
	WriterStateInit    WriterState = "init"
	WriterStateHeaders WriterState = "headers"
	WriterStateBody    WriterState = "body"
	WriterStateDone    WriterState = "done"
)

func NewResponseWriter(w io.Writer) *Writer {
	return &Writer{
		bw:    bufio.NewWriter(w),
		State: WriterStateInit,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	switch statusCode {
	case OK:
		w.StartLine = "HTTP/1.1 200 OK\r\n"
	case BadRequest:
		w.StartLine = "HTTP/1.1 400 Bad Request\r\n"
	case NotFound:
		w.StartLine = "HTTP/1.1 404 Not Found\r\n"
	case InternalError:
		w.StartLine = "HTTP/1.1 500 Internal Server Error\r\n"
	default:
		return fmt.Errorf("unsupported status code: %d", statusCode)
	}

	if _, err := w.bw.WriteString(w.StartLine); err != nil {
		return fmt.Errorf("error writing status line: %v", err)
	}

	w.State = WriterStateHeaders

	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.State != WriterStateHeaders {
		return fmt.Errorf("cannot write headers in state: %s", w.State)
	}

	w.Headers = headers

	var headerLines string
	for key, value := range headers {
		headerLines += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	headerLines += "\r\n"

	if _, err := w.bw.WriteString(headerLines); err != nil {
		return fmt.Errorf("error writing headers: %v", err)
	}

	w.State = WriterStateBody

	return nil
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	if w.State != WriterStateBody {
		return 0, fmt.Errorf("cannot write body in state: %s", w.State)
	}

	w.Body = body

	if _, err := w.bw.Write(body); err != nil {
		return 0, fmt.Errorf("error writing body: %v", err)
	}

	w.State = WriterStateDone

	if err := w.bw.Flush(); err != nil {
		return 0, fmt.Errorf("error flushing buffer: %v", err)
	}

	return 0, nil
}
