package pusher

import (
	"net/http"

	"encoding/json"

	"container/list"

	"fmt"
	"sync"
)

//PushMessage contains the data which can be sent to the client
type PushMessage struct {
	Event   string
	Payload interface{}
}

//PushHandler contains a list of all the listening clients and implements http.Handler
type PushHandler struct {
	clients     *list.List
	clientsLock *sync.RWMutex
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

	h.clientsLock.Lock()
	e := h.clients.PushBack(clientChan)
	h.clientsLock.Unlock()

	//close the clientChan when the client connection goes down
	closeChan := notifier.CloseNotify()
	go func() {
		<-closeChan //block until connection is closed
		close(clientChan)
	}()

	//at the end of the function or on error remove the client
	defer func() {
		h.clientsLock.Lock()
		h.clients.Remove(e)
		h.clientsLock.Unlock()
	}()

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

		fmt.Fprintf(res, "data: %s\n\n", string(b))

		flusher.Flush()
	}

	//we only get here when the channel is closed. Right?
}

//Push pushes a new message to all connected clients
func (h *PushHandler) Push(msg PushMessage) {

	//we will only block when the push handler is writing to the client list. So multiple goroutines are allowed to read from h.clients
	h.clientsLock.RLock()
	defer h.clientsLock.RUnlock()

	for e := h.clients.Front(); e != nil; e = e.Next() {
		e.Value.(chan PushMessage) <- msg
	}
}

//NewHandler creates a new push handler
func NewHandler() *PushHandler {
	return &PushHandler{
		clients:     list.New(),
		clientsLock: &sync.RWMutex{},
	}
}
