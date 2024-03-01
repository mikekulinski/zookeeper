package main

import (
	"context"
	"fmt"
	"log"

	zkc "github.com/mikekulinski/zookeeper/pkg/client"
	pbzk "github.com/mikekulinski/zookeeper/proto"
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

	fmt.Println("Connected to Zookeeper")

	cResp, err := client.Create(context.TODO(), &pbzk.CreateRequest{
		Path: "/zoo",
		Data: []byte("Secrets hahahahaha!!"),
	})
	if err != nil {
		log.Fatal("Error creating znode: ", err)
	}
	fmt.Println(cResp)

	gResp, err := client.GetData(context.TODO(), &pbzk.GetDataRequest{
		Path: "/zoo",
	})
	if err != nil {
		log.Fatal("Error getting data:", err)
	}
	fmt.Printf("%s", gResp.GetData())
}
