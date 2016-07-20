package pusher_test

import (
	"testing"

	"github.com/Remcoman/pusher"

	"net/http"

	"sync"
)

func TestReceive(t *testing.T) {
	p := pusher.NewHandler()

	println("starting servre")

	g := sync.WaitGroup{}
	g.Add(2)

	go func() {
		http.HandleFunc("/stream", func(res http.ResponseWriter, req *http.Request) {
			p.ServeHTTP(res, req)
		})
		http.ListenAndServe("127.0.0.1:8080", nil)
	}()

	println("stopping server")

	//send a event and check if it is registered
	//http.NewRequest()
}
