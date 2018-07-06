package cube_executor

import (
	handler "cube_executor/test_cube"
	"encoding/json"
	cube_interface "github.com/akaumov/cube"
	"github.com/akaumov/nats-pool"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/nats-io/go-nats"
	"github.com/satori/go.uuid"
	"sync"
)

const Version = "1"

type LogMessageParams struct {
	Time       int64  `json:"time"`
	Id         string `json:"id"`
	Class      string `json:"class"`
	InstanceId string `json:"instanceId"`
	Level      string `json:"level"`
	Text       string `json:"text"`
}

type Cube struct {
	busAddress        string
	instanceId        string
	class             string
	inputChannelsMap  map[string]string
	outputChannelsMap map[string]string
	params            map[string]string
	paramsMutex       sync.RWMutex
	pool              *nats_pool.Pool
	handler           cube_interface.HandlerInterface
}

func NewCube(instanceId string, inputChannelsMap map[string]string, outputChannelsMap map[string]string, params map[string]string) *Cube {
	return &Cube{
		busAddress:        "nats://cubes-bus:4444",
		inputChannelsMap:  inputChannelsMap,
		outputChannelsMap: outputChannelsMap,
		params:            params,
		handler:           &handler.Handler{},
		paramsMutex:       sync.RWMutex{},
	}
}

func (c *Cube) GetParam(param string) string {
	c.paramsMutex.RLock()
	defer func() { c.paramsMutex.RUnlock() }()
	return c.params[param]
}

func (c *Cube) GetClass() string {
	return c.class
}

func (c *Cube) GetInstanceId() string {
	return c.instanceId
}

func (c *Cube) PublishMessage(toChannel string, message cube_interface.Message) error {
	connection, err := c.pool.Get()
	defer func() { c.pool.Put(connection) }()

	if err != nil {
		return nil
	}

	encodedMessage, err := json.Marshal(message)
	if err != nil {
		return nil
	}

	err = connection.Publish(toChannel, encodedMessage)
	return err
}

func (c *Cube) CallMethod(channel string, request cube_interface.Request, timeout time.Duration) (*cube_interface.Response, error) {
	connection, err := c.pool.Get()
	defer func() { c.pool.Put(connection) }()

	if err != nil {
		return nil, err
	}

	encodedMessage, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	packedResponse, err := connection.Request(channel, encodedMessage, timeout)
	if err == nats.ErrTimeout {
		return nil, cube_interface.ErrorTimeout
	}

	var response cube_interface.Response
	err = json.Unmarshal(packedResponse.Data, &response)

	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Cube) sendLogMessage(level string, text string) error {

	id := uuid.NewV4().String()
	subject := "log." + c.class + "." + c.instanceId

	logMessage := LogMessageParams{
		Id:         id,
		Class:      c.class,
		InstanceId: c.instanceId,
		Time:       time.Now().UnixNano(),
		Level:      level,
		Text:       text,
	}

	message, _ := json.Marshal(logMessage)
	err := c.PublishMessage(subject, cube_interface.Message{
		Id:      &id,
		Version: Version,
		Params:  (*json.RawMessage)(&message),
		Method:  level,
	})

	return err
}

func (c *Cube) LogDebug(text string) error {
	return c.sendLogMessage("debug", text)
}

func (c *Cube) LogError(text string) error {
	return c.sendLogMessage("error", text)
}

func (c *Cube) LogFatal(text string) error {
	return c.sendLogMessage("fatal", text)
}

func (c *Cube) LogInfo(text string) error {
	return c.sendLogMessage("info", text)
}

func (c *Cube) LogWarning(text string) error {
	return c.sendLogMessage("warning", text)
}

func (c *Cube) LogTrace(text string) error {
	return c.sendLogMessage("trace", text)
}

func getOsSignalWatcher() chan os.Signal {

	stopChannel := make(chan os.Signal)
	signal.Notify(stopChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)

	return stopChannel
}

func (c *Cube) startListenMessagesFromBus(inputChannel string, stopChannel chan os.Signal) {

	busClient, err := c.pool.Get()
	defer func() { c.pool.Put(busClient) }()

	if err != nil {
		log.Fatalf("Can't connect to nats: %v", err)
		return
	}

	_, err = busClient.Subscribe(inputChannel, func(msg *nats.Msg) {
		var message cube_interface.Message

		err := json.Unmarshal(msg.Data, message)
		if err == nil {
			return
		}

		c.handler.OnReceiveMessage(c, inputChannel, message)
	})

	if err != nil {
		log.Fatalf("Can't connect to nats: %v", err)
		return
	}

	<-stopChannel
}

func (c *Cube) Start() {

	//TODO: check stopping
	stopSignal := getOsSignalWatcher()

	pool, err := nats_pool.New(c.busAddress, 10)
	if err != nil {
		log.Panicf("can't connect to nats: %v", err)
	}

	c.pool = pool
	defer func() { pool.Empty() }()

	for _, instanceChannel := range c.inputChannelsMap {
		go c.startListenMessagesFromBus(instanceChannel, getOsSignalWatcher())
	}

	func() {
		<-stopSignal
		c.handler.OnStop(c)
		c.Stop()
	}()
}

func (c *Cube) Stop() {

}

var _ cube_interface.Cube = (*Cube)(nil)
