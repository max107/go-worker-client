package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	zmq "github.com/pebbe/zmq4"
	"log"
	"net/http"
	"time"
)

const (
	SERVER_ENDPOINT = "tcp://192.168.0.89:5555"
)

type Command struct {
	Plugin string
	Action string
	Args   map[string]interface{}
}

func SendCommand(msg string) (string, error) {
	requester, _ := zmq.NewSocket(zmq.REQ)
	defer requester.Close()
	requester.Connect(SERVER_ENDPOINT)

	requester.Send(msg, 0)
	return requester.Recv(0)
}

func commandAction(plugin, action string, c *gin.Context) {
	cmd := Command{
		Plugin: plugin,
		Action: action,
	}

	binded := c.Bind(&cmd.Args)
	if !binded {
		log.Printf("Failed to bind")
	}

	msg, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
	}

	reply, err := SendCommand(string(msg))
	c.JSON(200, gin.H{"status": err != nil, "message": reply})
}

func main() {
	// msg := GetOpenvzMsg()

	r := gin.Default()

	// Global middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	v1 := r.Group("/v1")
	{
		v1.POST("/create/:plugin", func(c *gin.Context) {
			commandAction(c.Params.ByName("plugin"), "create", c)
		})
		v1.PUT("/update/:plugin", func(c *gin.Context) {
			commandAction(c.Params.ByName("plugin"), "update", c)
		})
		v1.DELETE("/delete/:plugin", func(c *gin.Context) {
			commandAction(c.Params.ByName("plugin"), "delete", c)
		})
	}

	s := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}
