package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	pb "github.com/daulet/stream/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type HttpHandler struct {
	client pb.TransformerClient
}

func NewHttpHandler(addr string) (*HttpHandler, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &HttpHandler{
		client: pb.NewTransformerClient(conn),
	}, nil
}

func (s *HttpHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-ndjson")

	stream, err := s.client.Generate(r.Context(), &pb.Request{})
	if err != nil {
		// TODO maybe error should be in message
		w.Write([]byte(fmt.Sprintf("error: %v", err)))
		return
	}

	enc := json.NewEncoder(w)
	f, _ := w.(http.Flusher)
	for {
		gen, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			w.Write([]byte(fmt.Sprintf("error: %v", err)))
			break
		}
		msg := &Message{
			Token: gen.Token,
		}
		if err := enc.Encode(msg); err != nil {
			w.Write([]byte(fmt.Sprintf("error: %v", err)))
			break
		}
		if f != nil {
			f.Flush()
		}
	}
}
