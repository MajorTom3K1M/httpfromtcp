package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(out)

		var str string
		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err != nil {
				break
			}

			data := buffer[:n]
			if i := bytes.IndexByte(data, byte('\n')); i != -1 {
				str += string(data[:i])
				out <- str
				data = data[i+1:]
				str = ""
			}

			str += string(data)
		}

		if len(str) != 0 {
			out <- str
		}
	}()

	return out
}

func main() {
	l, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("unable to start server", err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("accepted connection from %s\n", conn.RemoteAddr())

		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Printf("%s\n", line)
		}
	}
}
