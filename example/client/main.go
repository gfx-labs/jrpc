package main

import (
	"log"
	"time"

	"gfx.cafe/open/jrpc"
)

func main() {
	client, err := jrpc.Dial("https://mainnet.boba.network/")
	if err != nil {
		panic(err)
	}
	var b string
	for {
		err = client.Do(nil, &b, "eth_blockNumber", nil)
		if err != nil {
			log.Println(err)
			continue
		}
		time.Sleep(1 * time.Second)
		log.Println(b)
	}

}
