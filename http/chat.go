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
		fmt.Fprint(p, "Found one! Say hi.")
		fmt.Fprint(c, "Found one! Say hi.")
		errc := make(chan error, 1)
		go copy(p, c, errc)
		go copy(c, p, errc)
		if err := <-errc; err != nil {
			log.Println(err)
		}
		p.done <- true
		c.done <- true
	}
}

func copy(w io.Writer, r io.Reader, errc chan<- error) {
	_, err := io.Copy(w, r)
	errc <- err
}

const rootHtml = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8" />
<script>

var input, output, websocket;

function showMessage(m) {
	var p = document.createElement("p");
	p.innerHTML = m;
	output.appendChild(p);
}

function onMessage(e) {
	showMessage(e.data);
}

function onClose() {
	showMessage("Connection closed.");
}

function sendMessage() {
	var m = input.value;
	input.value = "";
	websocket.send(m);
	showMessage(m);
}

function onKey(e) {
	if (e.keyCode == 13) {
		sendMessage();
	}
}

function init() {
	input = document.getElementById("input");
	input.addEventListener("keyup", onKey, false);

	output = document.getElementById("output");

	websocket = new WebSocket("ws://` + listenAddr + `/socket");
	websocket.onmessage = onMessage;
	websocket.onclose = onClose;
}

window.addEventListener("load", init, false);

</script>
</head>
<body>
<input id="input" type="text">
<div id="output"></div>
</body>
</html>
`
