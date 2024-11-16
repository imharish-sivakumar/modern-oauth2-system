package grpcserver

import (
	"net"

	"google.golang.org/grpc"

	"github.com/imharish-sivakumar/modern-oauth2-system/cisauth-proto/pb"
)

// GRPCServer implements gRPC server implementation for token service.
type GRPCServer struct {
	server             *grpc.Server
	port               string
	tokenServiceServer pb.TokenServiceServer
}

// NewGRPCServer is a constructor and returns a pointer to GRPCServer object.
func NewGRPCServer(port string, tokenServiceServer pb.TokenServiceServer) *GRPCServer {
	return &GRPCServer{
		port:               port,
		tokenServiceServer: tokenServiceServer,
	}
}

// ListenAndServe spin up the server on a given port.
func (s *GRPCServer) ListenAndServe() error {
	lis, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}

	s.server = grpc.NewServer()
	pb.RegisterTokenServiceServer(s.server, s.tokenServiceServer)

	if err = s.server.Serve(lis); err != nil {
		return err
	}

	return nil
}

// Stop is to stop grpc server.
func (s *GRPCServer) Stop() {
	s.server.Stop()
}
