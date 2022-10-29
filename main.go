package main

import (
	"github.com/fideloper/myproxy/reverseproxy"
	"github.com/gorilla/mux"
	"log"
	"os"
	"os/signal"
)

func main() {
	r := &reverseproxy.ReverseProxy{}

	// Handle URI /foo
	a := mux.NewRouter()
	a.Host("fid.dev").Path("/foo")
	r.AddTarget("http://localhost:8001", a)

	// Handle anything else
	r.AddTarget("http://localhost:8000", nil)

	// Listen for http://
	r.AddListener(":80")

	// Listen for https://
	r.AddListenerTLS(":443", "keys/fid.dev.pem", "keys/fid.dev-key.pem")

	if err := r.Start(); err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Graceful shutdown
	r.Stop()
}
