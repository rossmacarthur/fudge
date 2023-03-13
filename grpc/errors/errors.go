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
