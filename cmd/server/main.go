package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"

	"github.com/mikekulinski/zookeeper/pkg/server"
)

const (
	serverName = "Zookeeper"
)

func main() {
	zk := server.NewServer()
	err := rpc.RegisterName(serverName, zk)
	if err != nil {
		log.Fatal("register error:", err)
	}
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("listen error:", err)
	}

	fmt.Println("Listening on localhost:8080...")
	err = http.Serve(l, nil)
	if err != nil {
		log.Fatal("serve error:", err)
	}
}
