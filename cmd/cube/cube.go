package main

import (
	"cube_executor"
	"flag"
	"log"
)

func main() {

	flag.Parse()
	log.SetFlags(0)

	var mapChannels map[string]string
	cube := cube_executor.NewCube("1", mapChannels, mapChannels, mapChannels)
	cube.Start()
}
