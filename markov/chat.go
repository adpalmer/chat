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
	rootTemplate.Execute(w, listenAddr)
}

type Socket struct {
	io.Reader
	io.Writer
	done chan bool
}

var chain = NewChain(2) // 2-word prefixes

func socketHandler(ws *websocket.Conn) {
	r, w := io.Pipe()
	go func() {
		_, err := io.Copy(io.MultiWriter(w, chain), ws)
		w.CloseWithError(err)
	}()
	s := Socket{r, ws, make(chan bool)}
	go mux(s)
	<-s.done
}

var partner = make(chan Socket)

func mux(c Socket) {
	fmt.Fprint(c, "Waiting for a partner...")
	select {
	case partner <- c:
		// now handled by the other goroutine
	case p := <-partner:
		chat(p, c)
	case <-time.After(time.Second * 10):
		chat(markovBot(), c)
	}
}

func chat(a, b Socket) {
	fmt.Fprintln(a, "Found one! Say hi.")
	fmt.Fprintln(b, "Found one! Say hi.")
	errc := make(chan error, 1)
	go cp(a, b, errc)
	go cp(b, a, errc)
	if err := <-errc; err != nil {
		log.Println(err)
	}
	a.done <- true
	b.done <- true
}

func cp(w io.Writer, r io.Reader, errc chan<- error) {
	_, err := io.Copy(w, r)
	errc <- err
}

// markovBot returns a Socket that responds to any read by
// writing a generated sentence.
func markovBot() Socket {
	r, out := io.Pipe() // outgoing data
	in, w := io.Pipe()  // incoming data
	go func() {
		b := make([]byte, 1024) // scratch buffer
		for {
			if _, err := in.Read(b); err != nil {
				break
			}
			msg := chain.Generate(10) // at most 10 words
			if _, err := out.Write([]byte(msg)); err != nil {
				break
			}
		}
	}()
	done := make(chan bool)
	go func() {
		<-done    // block until we get the done signal and then close
		w.Close() // the pipe, causing the blocking Read above to fail
	}()
	return Socket{r, w, done}
}
