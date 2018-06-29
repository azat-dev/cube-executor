package handler

import (
	"github.com/akaumov/cube"
)

type Handler struct {
}

func (h *Handler) OnStart(cubeInstance cube.Cube) {
}

func (h *Handler) OnStop(cubeInstance cube.Cube) {

}

func (h *Handler) OnReceiveMessage(cubeInstance cube.Cube, message cube.Message) {
	cubeInstance.PublishMessage("out", cube.Message{
		Id:      message.Id,
		From:    message.To,
		To:      "",
		Method:  message.Method,
		Params:  message.Params,
		Version: message.Version,
	})
}

var _ cube.HandlerInterface = (*Handler)(nil)
