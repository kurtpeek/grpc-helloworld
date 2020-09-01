package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

var latency = 1 * time.Second

type server struct {
	helloworld.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	time.Sleep(latency)
	log.Printf("gRPC method received: %v", in.GetName())
	return &helloworld.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func TestHelloWorld(t *testing.T) {
	var testCases = []struct {
		name    string
		timeout time.Duration
		addr    string
	}{
		{"timedOut", 500 * time.Millisecond, ":9999"},
		{"notTimedOut", 2 * time.Second, ":4242"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lis, err := net.Listen("tcp", ":")
			require.NoError(t, err)

			s := grpc.NewServer()
			helloworld.RegisterGreeterServer(s, &server{})
			go s.Serve(lis)
			defer s.Stop()

			conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
			require.NoError(t, err)
			defer conn.Close()
			c := helloworld.NewGreeterClient(conn)

			httpServer := &http.Server{
				Addr: tc.addr,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					helloReply, err := c.SayHello(r.Context(), &helloworld.HelloRequest{Name: "John Doe"})
					if err != nil {
						log.Printf("SayHello error: %v", err)
					}
					fmt.Fprint(w, helloReply.GetMessage())
				}),
			}

			go func() {
				require.NoError(t, httpServer.ListenAndServe())
			}()

			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				<-time.After(tc.timeout)
				cancel()
			}()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://localhost%s/sayhello", tc.addr), nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)

			if tc.timeout < latency {
				assert.EqualError(t, err, fmt.Sprintf(`Get "http://localhost%s/sayhello": context canceled`, tc.addr))
			} else {
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				body, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)
				log.Printf("Client received message body: %s", body)
			}
		})
	}
}
