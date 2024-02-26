package main

import (
	"fmt"
	"log"

	zkc "github.com/mikekulinski/zookeeper/pkg/client"
	"github.com/mikekulinski/zookeeper/pkg/zookeeper"
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

	cResp, err := client.Create(&zookeeper.CreateReq{
		Path: "/zoo",
		Data: []byte("Secrets hahahahaha!!"),
	})
	if err != nil {
		log.Fatal("Error creating znode: ", err)
	}
	fmt.Println(cResp)

	gResp, err := client.GetData(&zookeeper.GetDataReq{
		Path: "/zoo",
	})
	if err != nil {
		log.Fatal("Error getting data:", err)
	}
	fmt.Printf("%s", gResp.Data)
}
