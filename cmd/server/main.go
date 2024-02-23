package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"

	"github.com/mikekulinski/zookeeper/pkg/server"
)

func main() {
	zk := server.NewServer()
	err := rpc.Register(zk)
	if err != nil {
		log.Fatal("register error:", err)
	}
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	err = http.Serve(l, nil)
	if err != nil {
		log.Fatal("serve error:", err)
	}
}
