package main

// bowser serves the html output in the browser.
// client open an websocket connection, and the server push
// the new changes, once there's a new activity in the working file.
type browser struct {
	port int
	file string
	// parseFunc() // mark with configuration.
}

// browser.Serve()  // serve connections
// browser.Watch()  // watch for file changes (maybe also holds the connections?).
// browser.Open()   // open localhost:port in browser (wait for teardown).
