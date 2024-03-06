package main

import (
	"context"
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
	ctx := context.Background()
	client, err := zkc.NewClient(ctx, serverAddress)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer client.Close()

	log.Println("Connected to Zookeeper")

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
			Message: &pbzk.ZookeeperRequest_Create{
				Create: &pbzk.CreateRequest{
					Path: "/zoo/giraffe",
					Data: []byte("More secrets"),
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
		{
			Message: &pbzk.ZookeeperRequest_GetData{
				GetData: &pbzk.GetDataRequest{
					Path: "/zoo/giraffe",
				},
			},
		},
	}

	waitc := make(chan struct{})
	go func() {
		for {
			resp, err := client.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a message : %v", err)
			}
			log.Println(resp)
		}
	}()
	for _, request := range requests {
		if err := client.Send(request); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
	err = client.Close()
	if err != nil {
		log.Fatal("failed to close the stream")
	}
	<-waitc
}
