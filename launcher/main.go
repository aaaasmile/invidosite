package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	serverurl := "localhost:5573"
	staticSiteDirSrv := http.Dir("../static/invido")

	log.Println("static blog dir", staticSiteDirSrv)
	http.Handle("/", http.StripPrefix("/", http.FileServer(staticSiteDirSrv)))
	log.Println("Try this url for Blog: ", fmt.Sprintf("http://%s", serverurl))

	srv := &http.Server{
		Addr: serverurl,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      nil,
	}
	go func() {
		log.Println("start listening web with http")
		if err := srv.ListenAndServe(); err != nil {
			log.Println("Server is not listening anymore: ", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	log.Println("Enter in server blocking loop")

loop:
	for {
		select {
		case <-sig:
			log.Println("stop because interrupt")
			break loop
		}
	}

	log.Println("Bye, service")
}
