package cube_executor

import (
	"github.com/akaumov/nats-pool"
	cube_template "github.com/akaumov/cube"
	handler "cube_executor/test_cube"
	"os"
	"os/signal"
	"syscall"
	"log"
	"github.com/nats-io/go-nats"
	"time"
	"encoding/json"
)

type Cube struct {
	busAddress        string
	id                string
	inputChannelsMap  map[string]string
	outputChannelsMap map[string]string
	params            map[string]string
	pool              *nats_pool.Pool
	handler           cube_template.HandlerInterface
}

func NewCube(id string, inputChannelsMap map[string]string, outputChannelsMap map[string]string, params map[string]string) *Cube {
	return &Cube{
		busAddress:        "nats://localhost:4444",
		inputChannelsMap:  inputChannelsMap,
		outputChannelsMap: outputChannelsMap,
		params:            params,
		handler:           &handler.Handler{},
	}
}

func (c *Cube) GetParams() map[string]string {
	return c.params
}

func (c *Cube) GetId() string {
	return c.id
}

func (c *Cube) PublishMessage(toChannel string, message cube_template.Message) {

}

func (c *Cube) MakeRequest(channel string, message cube_template.Message, timeout time.Duration) {

}

func (c *Cube) Log(text string) {

}

func getOsSignalWatcher() chan os.Signal {

	stopChannel := make(chan os.Signal)
	signal.Notify(stopChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)

	return stopChannel
}

func (c *Cube) startListenMessagesFromBus(stopChannel chan os.Signal) {
	//TODO: Workers?
	busClient, err := c.pool.Get()
	if err != nil {
		log.Fatalf("Can't connect to nats: %v", err)
		return
	}

	_, err = busClient.Subscribe("FOO", func(msg *nats.Msg) {
		var message cube_template.Message

		err := json.Unmarshal(msg.Data, message)
		if err == nil {
			return
		}

		c.handler.OnReceiveMessage(c, message)
	})

	if err != nil {
		log.Fatalf("Can't connect to nats: %v", err)
		return
	}

	<-stopChannel
}

func (c *Cube) Start() {

	stopSignal := getOsSignalWatcher()

	pool, err := nats_pool.New(c.busAddress, 10)
	if err != nil {
		log.Panicf("can't connect to nats: %v", err)
	}

	c.pool = pool
	defer func() { pool.Empty() }()

	go func() {
		<-stopSignal
		c.handler.OnStop(c)
		c.Stop()
	}()

	c.startListenMessagesFromBus(getOsSignalWatcher())
}

func (c *Cube) Stop() {

}

var _ cube_template.Cube = (*Cube)(nil)
