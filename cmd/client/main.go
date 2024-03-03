package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

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

	fmt.Println("Connected to Zookeeper")

	requests := []*pbzk.ZookeeperRequest{
		{
			Message: &pbzk.ZookeeperRequest_Create{
				Create: &pbzk.CreateRequest{
					Path: "/zoo",
					Data: []byte("Secrets hahahahaha!!"),
				},
			},
		},
		{
			Message: &pbzk.ZookeeperRequest_GetData{
				GetData: &pbzk.GetDataRequest{
					Path: "/zoo",
				},
			},
		},
	}

	stream, err := client.Message(context.Background())
	if err != nil {
		log.Fatal("error initializing the stream with the server")
	}
	waitc := make(chan struct{})
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a message : %v", err)
			}
			fmt.Println(resp)
		}
	}()
	for _, request := range requests {
		if err := stream.Send(request); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
	err = stream.CloseSend()
	if err != nil {
		log.Fatal("failed to close the stream")
	}
	<-waitc
}
