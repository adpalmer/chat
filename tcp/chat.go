package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

const listenAddr = "localhost:4000"

func main() {
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go serve(c)
	}
}

var partner = make(chan io.ReadWriteCloser)

func serve(c io.ReadWriteCloser) {
	fmt.Fprint(c, "Waiting for a partner...")
	select {
	case partner <- c:
		// now handled by the other goroutine
	case p := <-partner:
		fmt.Fprintln(p, "Found one! Say hi.")
		fmt.Fprintln(c, "Found one! Say hi.")
		errc := make(chan error, 1)
		go cp(p, c, errc)
		go cp(c, p, errc)
		if err := <-errc; err != nil {
			log.Println(err)
		}
		p.Close()
		c.Close()
	}
}

func cp(w io.Writer, r io.Reader, errc chan<- error) {
	_, err := io.Copy(w, r)
	errc <- err
}
