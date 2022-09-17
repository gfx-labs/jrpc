package main

import (
	"log"
	"time"

	"gfx.cafe/open/jrpc"
)

func main() {
	client, err := jrpc.Dial("ws://localhost:8545")
	if err != nil {
		panic(err)
	}
	var b string
	for {
		err = client.Call(&b, "eth_blockNumber")
		if err != nil {
			log.Println(err)
			continue
		}
		time.Sleep(1 * time.Second)
		log.Println(b)
	}

}
