package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

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
	io.Reader
	io.WriteCloser
	done chan bool
}

func socketHandler(ws *websocket.Conn) {
	r, w := io.Pipe()
	go func() {
		_, err := io.Copy(io.MultiWriter(w, markov), ws)
		w.CloseWithError(err)
	}()
	s := Socket{r, ws, make(chan bool)}
	go serve(s)
	<-s.done
}

var (
	partner = make(chan Socket)
	markov  = NewChain(3)
)

const (
	markovDelay = 10 * time.Second
	markovWords = 10
)

func serve(c Socket) {
	fmt.Fprint(c, "Waiting for a partner...")
	select {
	case partner <- c:
		// now handled by the other goroutine
	case p := <-partner:
		chat(p, c)
	case <-time.After(markovDelay):
		chat(markovBot(), c)
	}
}

func chat(p, c Socket) {
	fmt.Fprint(p, "Found one! Say hi.")
	fmt.Fprint(c, "Found one! Say hi.")
	errc := make(chan error, 1)
	go cp(p, c, errc)
	go cp(c, p, errc)
	if err := <-errc; err != nil {
		log.Println(err)
	}
	p.done <- true
	c.done <- true
}

func cp(w io.Writer, r io.Reader, errc chan<- error) {
	_, err := io.Copy(w, r)
	errc <- err
}

func markovBot() Socket {
	r, out := io.Pipe()
	in, w := io.Pipe()
	go func() {
		b := make([]byte, 1024)
		for {
			if _, err := in.Read(b); err != nil {
				break
			}
			msg := markov.Generate(markovWords)
			if _, err := out.Write([]byte(msg)); err != nil {
				break
			}
		}
	}()
	done := make(chan bool)
	go func() {
		<-done
		w.Close()
	}()
	return Socket{r, w, done}
}
