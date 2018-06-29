package cube

import (
	"time"
	"encoding/json"
)

type Cube interface {
	GetParams() map[string]string
	GetId() string
	PublishMessage(toChannel string, message Message)
	MakeRequest(channel string, message Message, timeout time.Duration)
	Log(text string)
}

type Message struct {
	Version string          `json:"version"`
	Id      string          `json:"id"`
	From    string          `json:"from"`
	To      string          `json:"to"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type HandlerInterface interface {
	OnStart(instance Cube)
	OnStop(instance Cube)
	OnReceiveMessage(instance Cube, message Message)
}
