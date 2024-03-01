package main

import (
	"log"
	"net"

	zookeeper "github.com/mikekulinski/zookeeper/pkg/server"
	pbzk "github.com/mikekulinski/zookeeper/proto"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	zk := zookeeper.NewServer()
	pbzk.RegisterZookeeperServer(s, zk)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
