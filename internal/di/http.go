package di

import (
	"log"
	"net/http"
)

func startHTTPServer(router http.Handler) {
	go func() {
		log.Println("server started on :8080")

		err := http.ListenAndServe(":8080", router)
		if err != nil {
			log.Fatal(err)
		}
	}()
}
