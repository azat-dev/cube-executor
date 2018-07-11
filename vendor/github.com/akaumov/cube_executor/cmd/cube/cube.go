package main

import (
	"cube_executor"
	"flag"
	"log"
)

func main() {
	var configPath = flag.String("config", "/config.json", "config path")

	flag.Parse()
	log.SetFlags(0)

	cube, err := cube_executor.NewCube(*configPath)
	if err != nil {
		log.Fatalf("can't init cube instance: %v/n", err)
	}

	cube.Start()
}
