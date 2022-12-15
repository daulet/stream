package server

import pb "github.com/daulet/stream/stream"

const (
	alphabet            = "abcdefghijklmnopqrstuvwxyz"
	defaultOutputLength = 100
)

type Grpc struct {
	pb.UnimplementedTransformerServer
}

var _ pb.TransformerServer = (*Grpc)(nil)

func (s *Grpc) Generate(_ *pb.Request, stream pb.Transformer_GenerateServer) error {
	for i := 0; i < defaultOutputLength; i++ {
		if err := stream.Send(&pb.Generation{
			Token: string(alphabet[i%len(alphabet)]),
		}); err != nil {
			return err
		}
	}
	return nil
}
