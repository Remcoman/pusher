# push it!

Simple EventSource (Server Sent Events api) for golang  

Example go code:

	// create a new handler
	push := pusher.NewHandler()

	go func() {
		for {
			var testPayload = struct {
				Msg string `json:"msg"`
			}{
				Msg: "hello from me",
			}

			push.Push(pusher.PushMessage{Payload : testPayload})

			time.Sleep(time.Second)
		}
	}()

	//create a server
	http.Handle("/stream", push)
	http.ListenAndServe(":8080", nil)

Example js code:

	var e = new EventSource("http://localhost:8080/stream")
	e.onmessage = function (data) {
		console.log(data);
	}