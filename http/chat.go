package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"code.google.com/p/go.net/websocket"
)

const listenAddr = "localhost:4000"

func main() {
	http.HandleFunc("/", rootHandler)
	http.Handle("/socket", websocket.Handler(socketHandler))
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(rootHtml))
}

type Socket struct {
	io.ReadWriteCloser
	done chan bool
}

func socketHandler(ws *websocket.Conn) {
	s := Socket{ws, make(chan bool)}
	go serve(s)
	<-s.done
}

var partner = make(chan Socket)

func serve(c Socket) {
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
		p.done <- true
		c.done <- true
	}
}

func cp(w io.Writer, r io.Reader, errc chan<- error) {
	_, err := io.Copy(w, r)
	errc <- err
}
