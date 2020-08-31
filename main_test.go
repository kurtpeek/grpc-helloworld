package main

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/status"
)

type server struct {
	helloworld.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	time.Sleep(1 * time.Second)
	log.Printf("Received: %v", in.GetName())
	return &helloworld.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func TestHelloWorld(t *testing.T) {
	lis, err := net.Listen("tcp", ":")
	require.NoError(t, err)

	s := grpc.NewServer()
	helloworld.RegisterGreeterServer(s, &server{})
	go s.Serve(lis)
	defer s.Stop()

	log.Println("Dialing gRPC server...")
	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
	require.NoError(t, err)
	defer conn.Close()
	c := helloworld.NewGreeterClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-time.After(500 * time.Millisecond)
		cancel()
	}()

	log.Println("Making gRPC request...")
	_, err = c.SayHello(ctx, &helloworld.HelloRequest{Name: "John Doe"})
	assert.Equal(t, codes.Canceled, status.Code(err))
	assert.EqualError(t, err, "rpc error: code = Canceled desc = context canceled")
}
