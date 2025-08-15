package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	l, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Fatal("unable to resolve address", err)
	}

	conn, err := net.DialUDP("udp", nil, l)
	if err != nil {
		log.Fatal("unable to dial UDP", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("error reading from UDP connection", err)
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Fatal("error writing to UDP connection", err)
		}
	}
}
