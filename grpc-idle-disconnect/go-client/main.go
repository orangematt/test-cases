package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/orangematt/test-cases/grpc-idle-disconnect/service"
)

func stream(ctx context.Context, conn *grpc.ClientConn) {
	client := service.NewMessageServiceClient(conn)
	stream, err := client.StreamUpdates(ctx, &emptypb.Empty{}) // opts...
	if err != nil {
		fmt.Fprintf(os.Stderr, "StreamUpdates gRPC call failed: %v\n", err)
		return
	}

	for {
		update, err := stream.Recv()
		if err == io.EOF {
			fmt.Fprintf(os.Stderr, "StreamUpdates EOF\n")
			return
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "StreamUpdates gRPC error: %v\n", err)
			return
		}

		fmt.Println(time.Now(), update)
	}
}

func main() {
	var wg sync.WaitGroup
	defer wg.Wait()

	var serverAddress string
	flag.StringVar(&serverAddress, "addr", "localhost:9999", "specify server address to connect to")
	flag.Parse()

	// Dial the server
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})
	conn, err := grpc.Dial(serverAddress,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(creds))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to %s: %v\n", serverAddress, err)
		os.Exit(1)
	}
	defer conn.Close()

	// Stream data from the server, encode it to JSON, and print to stdout
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		stream(ctx, conn)
	}()

	// Wait for shutdown signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-c:
	case <-ctx.Done():
	}
	signal.Stop(c)
}
