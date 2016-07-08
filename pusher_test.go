package pusher_test

import (
	"testing"

	"github.com/Remcoman/pusher"

	"net/http"
)

func TestReceive(t *testing.T) {
	p := pusher.NewHandler()

	http.Handle("/stream", p)
	http.ListenAndServe("127.0.0.1:8080", nil)

	//send a event and check if it is registered
	//http.NewRequest()
}
