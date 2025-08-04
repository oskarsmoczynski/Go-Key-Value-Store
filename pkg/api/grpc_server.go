package api

import (
	"context"
	"fmt"
	"net"

	"github.com/oskarsmoczynski/Go-Key-Value-Store/pkg/store"
	"github.com/oskarsmoczynski/Go-Key-Value-Store/proto/kvstore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	kvstore.UnimplementedKVStoreServer
	store  *store.Store
	server *grpc.Server
}

func NewGRPCServer(store *store.Store) *GRPCServer {
	return &GRPCServer{
		store:  store,
		server: grpc.NewServer(),
	}
}

func (s *GRPCServer) Start(port int) error {
	kvstore.RegisterKVStoreServer(s.server, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	fmt.Printf("gRPC server starting on port %d...\n", port)

	return s.server.Serve(lis)
}

func (s *GRPCServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

func (s *GRPCServer) Set(ctx context.Context, req *kvstore.SetRequest) (*kvstore.SetResponse, error) {
	if req.Key == "" {
		return &kvstore.SetResponse{
			Success: false,
			Error:   "key cannot be empty",
		}, status.Error(codes.InvalidArgument, "key cannot be empty")
	}

	var ttlSeconds uint64
	if req.TtlSeconds > 0 {
		ttlSeconds = uint64(req.TtlSeconds)
	}

	s.store.Set(req.Key, req.Value, ttlSeconds, true)

	return &kvstore.SetResponse{
		Success: true,
		Error:   "",
	}, nil
}

func (s *GRPCServer) Get(ctx context.Context, req *kvstore.GetRequest) (*kvstore.GetResponse, error) {
	if req.Key == "" {
		return &kvstore.GetResponse{
			Found: false,
			Value: "",
			Error: "key cannot be empty",
		}, status.Error(codes.InvalidArgument, "key cannot be empty")
	}

	value, found := s.store.Get(req.Key)

	return &kvstore.GetResponse{
		Found: found,
		Value: value,
		Error: "",
	}, nil
}

func (s *GRPCServer) Delete(ctx context.Context, req *kvstore.DeleteRequest) (*kvstore.DeleteResponse, error) {
	if req.Key == "" {
		return &kvstore.DeleteResponse{
			Success: false,
			Error:   "key cannot be empty",
		}, status.Error(codes.InvalidArgument, "key cannot be empty")
	}

	s.store.Delete(req.Key)

	return &kvstore.DeleteResponse{
		Success: true,
		Error:   "",
	}, nil
}
