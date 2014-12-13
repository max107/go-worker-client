//
//  Hello World client.
//  Connects REQ socket to tcp://localhost:5555
//  Sends "Hello" to server, expects "World" back
//

package main

import (
	zmq "github.com/pebbe/zmq4"

	"encoding/json"
	"errors"
	"fmt"
	"log"
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

func GetMysqlMsg() string {
	args := make(map[string]interface{})
	args["database"] = "mimictl_user1"
	args["username"] = "mimictl_user1"
	args["password"] = "123456"
	cmd := &Command{Plugin: "mysql", Args: args}

	msg, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
	}
	return string(msg)
}

func GetPgsqlMsg() string {
	args := make(map[string]interface{})
	args["database"] = "mimictl_user"
	args["username"] = "mimictl_user"
	args["password"] = "123456"
	cmd := &Command{Plugin: "pgsql", Args: args}

	msg, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
	}
	return string(msg)
}

func SendCommand(server string, timeout time.Duration, retries int, msg string) error {
	log.Println("I: connecting to server...")
	client, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		panic(err)
	}
	client.Connect(server)

	poller := zmq.NewPoller()
	poller.Add(client, zmq.POLLIN)

	sequence := 0
	retriesLeft := retries
	for retriesLeft > 0 {
		//  We send a request, then we work to get a reply
		sequence++
		client.SendMessage(msg)

		for expect_reply := true; expect_reply; {
			//  Poll socket for a reply, with timeout
			sockets, err := poller.Poll(timeout)
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
					errors.New(fmt.Sprintf("E: malformed reply from server: %s\n", reply))
				} else {
					log.Printf("I: server replied OK (%s)\n", reply[0])
					retriesLeft = retries
					expect_reply = false
					return nil
				}
			} else {
				retriesLeft--
				if retriesLeft == 0 {
					errors.New("E: server seems to be offline, abandoning")
				} else {
					log.Println("W: no response from server, retrying...")
					//  Old socket is confused; close it and open a new one
					client.Close()
					client, _ = zmq.NewSocket(zmq.REQ)
					client.Connect(server)
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

	return nil
}

func main() {
	msg := GetPgsqlMsg()
	err := SendCommand(SERVER_ENDPOINT, REQUEST_TIMEOUT, REQUEST_RETRIES, msg)
	if err != nil {
		log.Printf("Fail")
	}
}
