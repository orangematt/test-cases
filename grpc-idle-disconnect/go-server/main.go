package main

import (
	"context"
	"flag"
	"fmt"
	"net"
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

type MessageServer struct {
	service.UnimplementedMessageServiceServer

	server *grpc.Server

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (s *MessageServer) StreamUpdates(
	_ *emptypb.Empty,
	stream service.MessageService_StreamUpdatesServer,
) error {
	fmt.Fprintf(os.Stderr, "%s -> new StreamUpdates request\n", time.Now())
	defer func() {
		fmt.Fprintf(os.Stderr, "%s -> StreamUpdates completed\n", time.Now())
	}()

	// Send an initial message on the stream
	m := &service.Message{
		S: "Hello",
	}
	if err := stream.Send(m); err != nil {
		return err
	}

	// We never send anything ever again. Just wait for either client
	// disconnect or server stop

	select {
	case <-stream.Context().Done():
		fmt.Fprintf(os.Stderr, "%s -> StreamUpdates context done\n", time.Now())
		return nil
	case <-s.ctx.Done():
		fmt.Fprintf(os.Stderr, "%s -> StreamUpdates server context done\n", time.Now())
		return nil
	}
}

func (s *MessageServer) Stop() {
	s.server.Stop()
	s.cancel()
	s.wg.Wait()
}

func runServer(address, keyFile, certFile string) (*MessageServer, error) {
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	grpcServiceServer := &MessageServer{
		server: grpcServer,
	}
	service.RegisterMessageServiceServer(grpcServer, grpcServiceServer)

	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	grpcServiceServer.ctx, grpcServiceServer.cancel =
		context.WithCancel(context.Background())

	grpcServiceServer.wg.Add(1)
	go func() {
		defer grpcServiceServer.wg.Done()
		_ = grpcServer.Serve(l)
	}()

	return grpcServiceServer, nil
}

func main() {
	var (
		address           string
		keyFile, certFile string
	)

	flag.StringVar(&address, "address", ":9999", "listening address")
	flag.StringVar(&keyFile, "key", "./service.pem", "path to TLS key file")
	flag.StringVar(&certFile, "cert", "./service.pem", "path to TLS certification file")
	flag.Parse()

	s, err := runServer(address, keyFile, certFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "runServer: %v\n", err)
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	signal.Stop(c)

	s.Stop()
}
