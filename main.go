//
//  Hello World client.
//  Connects REQ socket to tcp://localhost:5555
//  Sends "Hello" to server, expects "World" back
//

package main

import (
	zmq "github.com/pebbe/zmq4"

	"encoding/json"
	"fmt"
	"time"
)

const (
	REQUEST_TIMEOUT = 2500 * time.Millisecond //  msecs, (> 1000!)
	REQUEST_RETRIES = 3                       //  Before we abandon
	SERVER_ENDPOINT = "tcp://localhost:5555"
)

type Command struct {
	Plugin string
	Args   map[string]interface{}
}

type Output interface{}

func GetMsg() string {
	args := make(map[string]interface{})
	args["test"] = []string{"foo", "bar"}
	cmd := &Command{Plugin: "mysql", Args: args}

	msg, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
	}
	return string(msg)
}

func main() {
	fmt.Println("I: connecting to server...")
	client, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		panic(err)
	}
	client.Connect(SERVER_ENDPOINT)

	msg := GetMsg()

	poller := zmq.NewPoller()
	poller.Add(client, zmq.POLLIN)

	sequence := 0
	retries_left := REQUEST_RETRIES
	for retries_left > 0 {
		//  We send a request, then we work to get a reply
		sequence++
		client.SendMessage(msg)

		for expect_reply := true; expect_reply; {
			//  Poll socket for a reply, with timeout
			sockets, err := poller.Poll(REQUEST_TIMEOUT)
			if err != nil {
				break //  Interrupted
			}

			//  Here we process a server reply and exit our loop if the
			//  reply is valid. If we didn't a reply we close the client
			//  socket and resend the request. We try a number of times
			//  before finally abandoning:

			if len(sockets) > 0 {
				//  We got a reply from the server, must match sequence
				reply, err := client.RecvMessage(0)
				if err != nil {
					break //  Interrupted
				}

				var out Output
				if err := json.Unmarshal([]byte(msg), &out); err != nil {
					fmt.Printf("E: malformed reply from server: %s\n", reply)
				} else {
					fmt.Printf("I: server replied OK (%s)\n", reply[0])
					retries_left = REQUEST_RETRIES
					expect_reply = false
				}
			} else {
				retries_left--
				if retries_left == 0 {
					fmt.Println("E: server seems to be offline, abandoning")
					break
				} else {
					fmt.Println("W: no response from server, retrying...")
					//  Old socket is confused; close it and open a new one
					client.Close()
					client, _ = zmq.NewSocket(zmq.REQ)
					client.Connect(SERVER_ENDPOINT)
					// Recreate poller for new client
					poller = zmq.NewPoller()
					poller.Add(client, zmq.POLLIN)
					//  Send request again, on new socket
					client.SendMessage(msg)
				}
			}
		}
	}
	client.Close()
}
