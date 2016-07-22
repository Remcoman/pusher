package pusher_test

import (
	"bytes"
	"testing"

	"net"
	"net/http"

	"bufio"
	"time"

	"strings"

	"encoding/json"

	"github.com/Remcoman/pusher"
)

func closeableServer(handler http.Handler) (quit chan bool, err error) {
	s := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: handler,
	}
	l, err := net.Listen("tcp", ":8080")

	quit = make(chan bool)

	//listen for new connections
	go s.Serve(l)

	//listen for the quit message and then close the listener
	go func() {
		select {
		case <-quit:
			l.Close()
			break
		}
	}()

	return
}

func TestReceive(t *testing.T) {
	push := pusher.NewHandler()

	mux := http.NewServeMux()
	mux.Handle("/stream", push)

	_, err := closeableServer(mux)
	if err != nil {
		t.Fatal("Could not start up server")
	}

	done := make(chan bool)

	var testPayload = struct {
		Msg string `json:"msg"`
	}{
		Msg: "hello from me",
	}

	go func() {
		//create a connection
		conn, err := net.Dial("tcp", "127.0.0.1:8080")
		if err != nil {
			println("error!")
		}

		//write request headers
		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/stream", nil)
		req.Write(conn)

		//we need to make sure that the connection has been established
		time.Sleep(time.Second)

		//pusher should have a new connection to which messages can be published
		push.Push(pusher.PushMessage{Payload: testPayload})

		//time to read the response
		res, err := http.ReadResponse(bufio.NewReader(conn), nil)

		s := bufio.NewScanner(res.Body)
		s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 { //no data left
				return 0, nil, nil
			}

			if i := bytes.Index(data, []byte("\n\n")); i >= 0 {
				return i + 2, data[0:i], nil
			}
			// If we're at EOF, we have a final, non-terminated line. Return it.
			if atEOF {
				return len(data), data, nil
			}
			// Request more data.
			return 0, nil, nil
		})

		if s.Scan() {
			parts := strings.Split(s.Text(), ": ")

			var dst map[string]interface{}

			err := json.Unmarshal([]byte(parts[1]), &dst)
			if err != nil {
				t.Log(err)
				t.Error("Got invalid json!")
			} else {
				t.Logf("Got valid data: %s", parts[1])
			}
		} else {
			t.Error("Could not find any data!")
		}

		done <- true //signal that we are done
	}()

	<-done
}
