package main

import (
	"fmt"
	"log"
	"net/rpc"

	"github.com/mikekulinski/zookeeper/pkg/zookeeper"
)

const (
	serverAddress = "localhost"
)

func main() {
	client, err := rpc.DialHTTP("tcp", serverAddress+":8080")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	// Synchronous call
	req := &zookeeper.CreateReq{
		Path:  "/x/p",
		Data:  []byte("Hello World!"),
		Flags: []zookeeper.Flag{zookeeper.SEQUENTIAL},
	}
	reply := &zookeeper.CreateResp{}
	err = client.Call("Server.Create", req, reply)
	if err != nil {
		log.Fatal("server error:", err)
	}
	fmt.Println(reply.ZNodeName)
}
