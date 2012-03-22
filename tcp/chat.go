package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp", ":4000")
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
		fmt.Fprint(p, "found one! Say hi.\n")
		fmt.Fprint(c, "found one! Say hi.\n")
		errc := make(chan error, 1)
		go copy(p, c, errc)
		go copy(c, p, errc)
		if err := <-errc; err != nil {
			log.Println(err)
		}
		p.Close()
		c.Close()
	}
}

func copy(w io.Writer, r io.Reader, errc chan<- error) {
	_, err := io.Copy(w, r)
	errc <- err
}
