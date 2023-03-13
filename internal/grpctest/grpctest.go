package grpctest

import (
	"context"

	"github.com/rossmacarthur/fudge/errors"
	grpcerrors "github.com/rossmacarthur/fudge/grpc/errors"
	"github.com/rossmacarthur/fudge/internal/grpctest/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ############################################################################
// Server
// ############################################################################

var _ pb.CandyStoreServer = (*Server)(nil)

type Server struct {
	pb.UnsafeCandyStoreServer
	client *Client
	errFn  func() error
}

func NewServer(addr string) (*Server, error) {
	client, err := NewClient(addr,
		grpc.WithUnaryInterceptor(grpcerrors.UnaryClientInterceptor))
	if err != nil {
		return nil, err
	}

	return &Server{client: client}, nil
}

func (s *Server) SetErrFn(fn func() error) {
	s.errFn = fn
}

func (s *Server) Buy(ctx context.Context, req *pb.BuyRequest) (*emptypb.Empty, error) {
	if req.Hops > 0 {
		err := s.client.Buy(ctx, req.Hops-1)
		if err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}

	var err error
	if s.errFn != nil {
		err = s.errFn()
	}
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	return &emptypb.Empty{}, nil
}

// ############################################################################
// Client
// ############################################################################

type Client struct {
	client pb.CandyStoreClient
}

func NewClient(addr string, opts ...grpc.DialOption) (*Client, error) {
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{pb.NewCandyStoreClient(conn)}, nil
}

func (c *Client) Buy(ctx context.Context, hops int64) error {
	_, err := c.client.Buy(ctx, &pb.BuyRequest{Hops: hops})
	return err
}
