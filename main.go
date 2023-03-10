package main

import (
	"k8s-webhook-validate/webhook"
	"log"
	"net/http"
	"time"
)

var (
	deploy = webhook.Deploy{}
)

var (
	tlsCrt = "config/tls/tls.crt"
	tlsKey = "config/tls/tls.key"
)

func main() {

	mux := http.NewServeMux()

	// exec deploy pod name renew
	mux.HandleFunc("/deployment/validating", deploy.Validating)

	server := &http.Server{
		Addr:        ":8443",
		Handler:     mux,
		ReadTimeout: 20 * time.Second, WriteTimeout: 20 * time.Second,
	}

	// add healthCheck
	go func() {
		healthCheck()
	}()

	log.Println("Validate http server start running on port :8443 ...")
	log.Fatal(server.ListenAndServeTLS(tlsCrt, tlsKey))

}

func healthCheck() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health_check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
