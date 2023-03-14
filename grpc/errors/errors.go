package grpcerrors

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryClientInterceptor is a gRPC client interceptor that converts gRPC status
// details into Fudge errors.
func UnaryClientInterceptor(ctx context.Context, method string, req, resp any,
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	err := invoker(ctx, method, req, resp, cc, opts...)
	return interceptClient(err)
}

// UnaryServerInterceptor is a gRPC server interceptor that returns Fudge error
// information in the gRPC status details.
func UnaryServerInterceptor(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

	resp, err := handler(ctx, req)
	return resp, interceptServer(err)
}

func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string, streamer grpc.Streamer,
	opts ...grpc.CallOption) (grpc.ClientStream, error) {

	cs, err := streamer(ctx, desc, cc, method, opts...)
	return &clientStream{ClientStream: cs}, interceptServer(err)
}

type clientStream struct {
	grpc.ClientStream
}

func (s *clientStream) SendMsg(m any) error {
	return interceptClient(s.ClientStream.SendMsg(m))
}

func (s *clientStream) RecvMsg(m any) error {
	return interceptClient(s.ClientStream.RecvMsg(m))
}

func StreamServerInterceptor(srv any, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	return interceptServer(handler(srv, ss))
}
