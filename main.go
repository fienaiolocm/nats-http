package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/nats-io/nats.go"
)

func main() {
	h1 := func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("req is %+v\n", req)
		io.WriteString(w, fmt.Sprintf("Hello from a H1 for %q!\n", req.URL.Path))
	}

	// could be nats.HandleFunc() if added.
	natsHandleFunc("foo", h1)

	// Handle via HTTP
	http.HandleFunc("/foo", h1)

	log.Printf("Listening on HTTP localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func natsConnect() *nats.Conn {
	if nc, err := nats.Connect("localhost"); err == nil {
		log.Printf("NATS connected to localhost")
		return nc
	} else if nc, err := nats.Connect("demo.nats.io"); err == nil {
		log.Printf("NATS connected to demo.nats.io")
		return nc
	}
	log.Fatalf("Could not connect to NATS System")
	return nil
}

var nc *nats.Conn

func natsHandleFunc(subject string, handler func(http.ResponseWriter, *http.Request)) {
	// NATS Setup
	if nc == nil {
		nc = natsConnect()
	}
	var _rb [256]byte
	respBuf := bytes.NewBuffer(_rb[:0])

	_, err := nc.Subscribe(subject, func(m *nats.Msg) {
		// Determine if HTTP request format. For now assume its not and construct one.
		buf := bytes.NewBuffer(m.Data)
		req, err := http.NewRequest("GET", subject, buf)
		if err != nil {
			log.Printf("Error creating http request: %v", err)
		}
		rr := httptest.NewRecorder()
		// Call into our handler.
		handler(rr, req)
		rr.Result().Write(respBuf)
		m.Respond(respBuf.Bytes())
	})

	if err != nil {
		log.Fatalf("NATS Error subscribing to %q, %v", subject, err)
	}
}
