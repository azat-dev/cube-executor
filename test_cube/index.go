package test_cube

import (
	cube_interface "github.com/akaumov/cube"
)

type Handler struct {
}

func (h *Handler) OnStart(instance cube_interface.Cube) {
	instance.LogInfo("OnStart")
}

func (h *Handler) OnStop(instance cube_interface.Cube) {
	instance.LogInfo("OnStop")
}

func (h *Handler) OnReceiveMessage(instance cube_interface.Cube, channel string, message cube_interface.Message) {
	instance.LogInfo("OnReceiveMessage")
}

func (h *Handler) OnReceiveRequest(instance cube_interface.Cube, channel string, request cube_interface.Request) (*cube_interface.Response, error) {
	instance.LogInfo("OnReceiveRequest")

	return &cube_interface.Response{
		Version: "1",
		Errors:  nil,
		Result:  nil,
	}, nil
}

var _ cube_interface.HandlerInterface = (*Handler)(nil)
