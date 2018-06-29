package main

import "cube_executor"

func main()  {
	//pool, err := nats_pool.New("nats://bus:4222", 10)
	//if err != nil {
	//	// handle error
	//}
	//
	//natsClient, err := pool.Get()
	//if err != nil {
	//	log.Fatalf("can't  connect to bus:\n%v", err)
	//	return
	//}
	//
	//timeout := time.Duration(100) * time.Millisecond
	//const message = `
	//	"method": "getServiceParams",
	//	"params": {
	//
	//`
	//natsClient.Request("_SYSTEM", []byte("{}"), timeout)

	var mapChannels map[string]string
	cube := cube_executor.NewCube("1", mapChannels, mapChannels, mapChannels)
	cube.Start()
}