package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/daulet/stream/server"
	pb "github.com/daulet/stream/stream"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := stream(); err != nil {
		panic(err)
	}
}

func stream() error {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := grpc.NewServer()
	defer srv.Stop()

	wg.Add(1)
	go func() {
		defer wg.Done()

		lstr, err := net.Listen("tcp", ":8080")
		if err != nil {
			fmt.Printf("failed to listen: %v\n", err)
			return
		}

		pb.RegisterTransformerServer(srv, &server.Grpc{})
		if err := srv.Serve(lstr); err != nil {
			fmt.Printf("failed to serve: %v\n", err)
		}
	}()

	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	client := pb.NewTransformerClient(conn)
	stream, err := client.Generate(ctx, &pb.Request{})
	if err != nil {
		return err
	}
	for {
		gen, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf("received: %s\n", gen.Token)
		<-time.After(20 * time.Millisecond)
	}

	return nil
}
