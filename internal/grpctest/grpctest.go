package grpctest

import (
	"context"
	"io"

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

var inStock = []string{"chocolate", "gummy", "lollipop"}

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

func (s *Server) Buy(ctx context.Context, req *pb.BuyRequest) (*pb.Candy, error) {
	if req.Hops > 0 {
		err := s.client.Buy(ctx, req.Hops-1)
		if err != nil {
			return nil, err
		}
		return &pb.Candy{Name: "chocolate"}, nil
	}

	err := s.maybeError()
	if err != nil {
		return nil, err
	}

	return &pb.Candy{Name: "chocolate"}, nil
}

func (s *Server) StreamCandyTo(stream pb.CandyStore_StreamCandyToServer) error {
	for {
		candy, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		for _, name := range inStock {
			if name == candy.Name {
				return errors.New("already in stock")
			}
		}
	}

	err := s.maybeError()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) StreamCandyFrom(_ *emptypb.Empty, stream pb.CandyStore_StreamCandyFromServer) error {
	for _, name := range inStock {
		err := stream.Send(&pb.Candy{Name: name})
		if err != nil {
			return err
		}
	}

	err := s.maybeError()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) maybeError() error {
	var err error
	if s.errFn != nil {
		err = s.errFn()
	}
	if err != nil {
		return errors.Wrap(err, "")
	}

	return nil
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

func (c *Client) StreamCandyTo(ctx context.Context, candy []string) error {
	stream, err := c.client.StreamCandyTo(ctx)
	if err != nil {
		return err
	}

	for _, name := range candy {
		err := stream.Send(&pb.Candy{Name: name})
		if err != nil {
			return err
		}
	}

	_, err = stream.CloseAndRecv()
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}

func (c *Client) StreamCandyFrom(ctx context.Context) ([]string, error) {
	stream, err := c.client.StreamCandyFrom(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var candy []string

	for {
		c, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}

		candy = append(candy, c.Name)
	}

	return candy, nil
}
