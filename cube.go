package cube_executor

import (
	handler "cube_executor/test_cube"
	"encoding/json"
	"fmt"
	cube_interface "github.com/akaumov/cube"
	"github.com/akaumov/nats-pool"
	"github.com/nats-io/go-nats"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const Version = "1"

type HostPort uint
type CubePort uint
type Protocol string

type CubeChannel cube_interface.Channel
type BusChannel cube_interface.Channel

type PortMap struct {
	CubePort CubePort `json:"cubePort"`
	HostPort HostPort `json:"hostPort"`
	Protocol Protocol `json:"protocol"`
}

type CubeConfig struct {
	SchemaVersion     string                     `json:"schemaVersion"`
	Version           string                     `json:"version"`
	Name              string                     `json:"name"`
	Source            string                     `json:"source"`
	Params            map[string]string          `json:"params"`
	PortsMapping      []PortMap                  `json:"portsMapping"`
	ChannelsMapping   map[CubeChannel]BusChannel `json:"channelsMapping"`
	NumberOfListeners int                        `json:"numberOfListeners"`
}

type LogMessageParams struct {
	Time       int64  `json:"time"`
	Id         string `json:"id"`
	Class      string `json:"class"`
	InstanceId string `json:"instanceId"`
	Level      string `json:"level"`
	Text       string `json:"text"`
}

type Cube struct {
	version             string
	busAddress          string
	instanceId          string
	class               string
	cubeChannelsMapping map[CubeChannel]BusChannel
	busChannelsMapping  map[BusChannel]CubeChannel
	inputChannels       []CubeChannel
	params              map[string]string
	paramsMutex         sync.RWMutex
	pool                *nats_pool.Pool
	handler             cube_interface.HandlerInterface
}

func NewCube() (*Cube, error) {
	configRaw, err := ioutil.ReadFile("/config.json")
	if err != nil {
		return nil, fmt.Errorf("can't read config file: %v/n", err)
	}

	var config CubeConfig
	err = json.Unmarshal(configRaw, &config)
	if err != nil {
		return nil, fmt.Errorf("can't parse config file: %v/n", err)
	}

	//TODO check channels mapping
	busChannelsMapping := map[BusChannel]CubeChannel{}

	for cubeChannel, busChannel := range config.ChannelsMapping {
		busChannelsMapping[busChannel] = cubeChannel
	}

	return &Cube{
		busAddress:          "nats://cubes-bus:4444",
		class:               config.Source,
		version:             config.Version,
		instanceId:          config.Name,
		cubeChannelsMapping: config.ChannelsMapping,
		busChannelsMapping:  busChannelsMapping,
		params:              config.Params,
		handler:             &handler.Handler{},
		inputChannels:       []CubeChannel{},
		paramsMutex:         sync.RWMutex{},
	}, nil
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

func (c *Cube) mapToBusChannel(channel CubeChannel) BusChannel {
	busChannel := c.cubeChannelsMapping[channel]
	if busChannel == BusChannel("") {
		busChannel = BusChannel(channel)
	}

	return busChannel
}

func (c *Cube) mapToCubeChannel(channel BusChannel) CubeChannel {
	cubeChannel := c.busChannelsMapping[channel]
	if cubeChannel == CubeChannel("") {
		cubeChannel = CubeChannel(channel)
	}

	return cubeChannel
}

func (c *Cube) PublishMessage(cubeChannel cube_interface.Channel, message cube_interface.Message) error {
	connection, err := c.pool.Get()
	defer func() { c.pool.Put(connection) }()

	if err != nil {
		return nil
	}

	encodedMessage, err := json.Marshal(message)
	if err != nil {
		return nil
	}

	busChannel := c.mapToBusChannel(CubeChannel(cubeChannel))
	err = connection.Publish(string(busChannel), encodedMessage)
	return err
}

func (c *Cube) CallMethod(cubeChannel cube_interface.Channel, request cube_interface.Request, timeout time.Duration) (*cube_interface.Response, error) {
	busChannel := c.mapToBusChannel(CubeChannel(cubeChannel))
	connection, err := c.pool.Get()
	defer func() { c.pool.Put(connection) }()

	if err != nil {
		return nil, err
	}

	encodedMessage, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	packedResponse, err := connection.Request(string(busChannel), encodedMessage, timeout)
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

	packedLogMessage, _ := json.Marshal(logMessage)

	message := cube_interface.Message{
		Id:      &id,
		Version: Version,
		Method:  level,
		Params:  (*json.RawMessage)(&packedLogMessage),
	}

	connection, err := c.pool.Get()
	defer func() { c.pool.Put(connection) }()

	if err != nil {
		return nil
	}

	encodedMessage, err := json.Marshal(message)
	if err != nil {
		return nil
	}

	err = connection.Publish(subject, encodedMessage)
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

func (c *Cube) startListenMessagesFromBus(inputChannel BusChannel, stopChannel chan os.Signal) {

	busClient, err := c.pool.Get()
	defer func() { c.pool.Put(busClient) }()

	if err != nil {
		log.Fatalf("Can't connect to nats: %v", err)
		return
	}

	_, err = busClient.Subscribe(string(inputChannel), func(msg *nats.Msg) {
		var message cube_interface.Message

		err := json.Unmarshal(msg.Data, message)
		if err == nil {
			return
		}

		cubeChannel := c.mapToCubeChannel(inputChannel)
		c.handler.OnReceiveMessage(c, cube_interface.Channel(cubeChannel), message)
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

	inputChannels := c.handler.OnInitInstance()

	for _, inputChannel := range inputChannels {
		busChannel := c.mapToBusChannel(CubeChannel(inputChannel))
		go c.startListenMessagesFromBus(busChannel, getOsSignalWatcher())
	}

	func() {
		<-stopSignal
		c.Stop()
	}()
}

func (c *Cube) Stop() {
	c.handler.OnStop(c)
}

var _ cube_interface.Cube = (*Cube)(nil)
