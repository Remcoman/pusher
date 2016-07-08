package pusher

import (
	"net/http"

	"encoding/json"

	"container/list"

	"fmt"
)

type PushMessage struct {
	Event   string
	Payload interface{}
}

type PushHandler struct {
	clients *list.List
}

func (h *PushHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/event-stream")
	res.Header().Add("Cache-Control", "no-cache")

	flusher, ok := res.(http.Flusher)
	if !ok {
		panic("Given server does not support flushing")
	}

	notifier, ok := res.(http.CloseNotifier)
	if !ok {
		panic("Given server does not notify on close")
	}

	clientChan := make(chan PushMessage, 0)

	e := h.clients.PushBack(list.Element{Value: clientChan})

	//close the clientChan when the client connection goes down
	closeChan := notifier.CloseNotify()
	go func() {
		<-closeChan //block until connection is closed
		close(clientChan)
	}()

	//at the end of the function or on error remove the client
	defer h.clients.Remove(e)

	//keep receiving data unless clientChan is closed
	for x := range clientChan {

		if x.Event != "" {
			fmt.Fprintf(res, "event: %s\n", x.Event)
		}

		b, err := json.Marshal(x.Payload)
		if err != nil {
			//todo implement logging
			continue
		}

		fmt.Fprint(res, "data: %s\n\n", string(b))

		flusher.Flush()
	}

	//we only get here when the channel is closed. Right?
}

func (h *PushHandler) Push(msg PushMessage) {
	for e := h.clients.Front(); e != nil; e = e.Next() {
		e.Value.(chan PushMessage) <- msg
	}
}

func NewHandler() *PushHandler {
	return &PushHandler{
		clients: list.New(),
	}
}
