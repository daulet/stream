package grpc

import pb "github.com/daulet/stream/stream"

type Server struct {
	pb.UnimplementedTransformerServer
}

var _ pb.TransformerServer = (*Server)(nil)

func (s *Server) Generate(*pb.Request, pb.Transformer_GenerateServer) error {
	panic("implement me")
}
