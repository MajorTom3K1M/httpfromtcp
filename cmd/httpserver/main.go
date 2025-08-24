package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const PORT = 42069

func main() {
	server, err := server.Serve(PORT, handler)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	log.Println("Server is running on port", PORT)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	path := req.RequestLine.RequestTarget
	switch {
	case path == "/yourproblem":
		handler400(w, req)
	case path == "/myproblem":
		handler500(w, req)
	case path == "/video":
		handlerVideo(w, req)
	case strings.HasPrefix(path, "/httpbin"):
		handlerChunked(w, req)
	default:
		handler200(w, req)
	}
}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.BadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.InternalError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.OK)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handlerChunked(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	resp, err := http.Get("https://httpbin.org/" + target)
	if err != nil {
		handler500(w, nil)
	}
	defer resp.Body.Close()

	hdrs := response.GetDefaultHeaders(0)
	hdrs.Del("Content-Length")
	hdrs.Override("Transfer-Encoding", "chunked")
	hdrs.Override("Trailer", "X-Content-Sha256, X-Content-Length")

	w.WriteStatusLine(response.OK)
	w.WriteHeaders(hdrs)

	var fullRespBody bytes.Buffer
	respBody := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(respBody)
		if n > 0 {
			w.WriteChunkedBody(respBody[:n])
			fullRespBody.Write(respBody[:n])
		}
		if err != nil {
			fmt.Printf("Finished streaming with error: %v\n", err)
			break
		}
	}

	w.WriteChunkedBodyDone()

	sha256Sum := sha256.Sum256(fullRespBody.Bytes())

	trailerHdrs := headers.Headers{
		"X-Content-Sha256": fmt.Sprintf("%x", sha256Sum),
		"X-Content-Length": fmt.Sprintf("%d", fullRespBody.Len()),
	}

	w.WriteTrailers(trailerHdrs)
}

func handlerVideo(w *response.Writer, req *request.Request) {
	w.WriteStatusLine(response.OK)

	body, err := os.ReadFile("./assets/vim.mp4")
	if err != nil {
		handler500(w, req)
		return
	}

	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "video/mp4")

	w.WriteHeaders(h)
	w.WriteBody(body)
}
