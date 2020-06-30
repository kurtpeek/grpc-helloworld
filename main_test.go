package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	port = ":50051"
)

type server struct {
	helloworld.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &helloworld.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func TestHelloWorld(t *testing.T) {
	lis, err := net.Listen("tcp", port)
	require.NoError(t, err)

	s := grpc.NewServer()
	helloworld.RegisterGreeterServer(s, &server{})
	go s.Serve(lis)
	defer s.Stop()

	log.Println("Dialing gRPC server...")
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%s", port), grpc.WithInsecure(), grpc.WithBlock())
	require.NoError(t, err)
	defer conn.Close()
	c := helloworld.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	log.Println("Making gRPC request...")
	r, err := c.SayHello(ctx, &helloworld.HelloRequest{Name: "John Doe"})
	require.NoError(t, err)
	log.Printf("Greeting: %s", r.GetMessage())
}
