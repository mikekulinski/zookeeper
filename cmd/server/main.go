package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		sig := <-sigCh
		log.Printf("got signal %v, attempting graceful shutdown", sig)
		s.GracefulStop()
		wg.Done()
	}()

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	wg.Wait()
	log.Println("clean shutdown")
}
