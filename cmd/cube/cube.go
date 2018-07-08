package main

import (
	"cube_executor"
	"flag"
	"log"
)

func main() {

	flag.Parse()
	log.SetFlags(0)

	cube, err := cube_executor.NewCube()
	if err != nil {
		log.Fatalf("can't init cube instance: %v/n", err)
	}

	cube.Start()
}
