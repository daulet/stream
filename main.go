package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	pb "github.com/daulet/stream/proto"
	"github.com/daulet/stream/server"

	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	serve := flag.Bool("serve", false, "whether program should serve an external request")
	flag.Parse()

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	if *serve {
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}

	shutdown := make(chan struct{})
	go func() {
		defer close(shutdown)
		if err := stream(ctx, *port); err != nil {
			panic(err)
		}
	}()

	if !*serve {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d", *port))
		if err != nil {
			panic(err)
		}
		dec := json.NewDecoder(resp.Body)
		defer resp.Body.Close()
		for {
			msg := &server.Message{}
			if err := dec.Decode(msg); err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}
			fmt.Println(msg.Token)
		}
		cancel()
	}

	<-shutdown
}

func stream(ctx context.Context, port int) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	grpcSrv := grpc.NewServer(
		grpc.ChainStreamInterceptor(server.StreamInterceptor()),
	)
	defer grpcSrv.Stop()

	wg.Add(1)
	go func() {
		defer wg.Done()

		lstr, err := net.Listen("tcp", fmt.Sprintf(":%d", port+1))
		if err != nil {
			fmt.Printf("failed to listen: %v\n", err)
			return
		}

		pb.RegisterTransformerServer(grpcSrv, &server.Grpc{})
		if err := grpcSrv.Serve(lstr); err != nil {
			fmt.Printf("failed to serve: %v\n", err)
		}
	}()

	handler, err := server.NewHttpHandler(fmt.Sprintf("localhost:%d", port+1))
	if err != nil {
		return err
	}
	http.HandleFunc("/", handler.Handle)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	defer srv.Shutdown(ctx)

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("failed to serve: %v\n", err)
		}
	}()

	for range ctx.Done() {
	}
	return nil
}
