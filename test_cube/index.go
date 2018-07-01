package test_cube

import (
	"github.com/akaumov/cube"
	"log"
)

type Handler struct {
}

func (h *Handler) OnStart(cube cube.Cube) {
	log.Println("OnStart")
}

func (h *Handler) OnStop(cube cube.Cube) {
	log.Println("OnStop")
}

func (h *Handler) OnReceiveMessage(cube cube.Cube, message cube.Message) {
	log.Println("OnStart", message)
}

var _ cube.HandlerInterface = (*Handler)(nil)
