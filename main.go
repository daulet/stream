package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"

	pb "github.com/daulet/stream/proto"
	"github.com/daulet/stream/server"

	"google.golang.org/grpc"
)

func main() {
	if err := stream(8080); err != nil {
		panic(err)
	}
}

func stream(port int) error {
	ctx := context.Background()

	var wg sync.WaitGroup
	defer wg.Wait()

	grpcSrv := grpc.NewServer()
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

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		return err
	}
	dec := json.NewDecoder(resp.Body)
	for {
		msg := &server.Message{}
		if err := dec.Decode(msg); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmt.Println(msg.Token)
	}
	return nil
}
