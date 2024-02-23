package main

import (
	"log"

	zkc "github.com/mikekulinski/zookeeper/pkg/client"
)

const (
	serverAddress = "localhost"
)

func main() {
	client, err := zkc.NewClient(serverAddress)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer func() {
		err := client.Close()
		if err != nil {
			log.Fatal("error closing client:", err)
		}
	}()
}
