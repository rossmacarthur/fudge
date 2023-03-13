package grpcerrors_test

import (
	"context"
	"strings"

	"fmt"
	"math/rand"
	"net"
	"testing"

	"github.com/rossmacarthur/fudge/errors"
	grpcerrors "github.com/rossmacarthur/fudge/grpc/errors"
	"github.com/rossmacarthur/fudge/internal/grpctest"
	"github.com/rossmacarthur/fudge/internal/grpctest/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestInterceptors(t *testing.T) {
	tests := []struct {
		name string

		// noServer doesn't run the gRPC server
		noServer bool

		// noServerIntercept doesn't add the server gRPC interceptor
		noServerIntercept bool

		// noClientIntercept doesn't add the client gRPC interceptor
		noClientIntercept bool

		// hops is the number of hops to make
		hops int64

		// errFn generates the error on the server
		errFn func() error

		// expFn asserts any conditions this test case requires
		expFn func(t *testing.T, got error)
	}{
		{
			name:              "no interceptor: nil",
			noServerIntercept: true,
			errFn: func() error {
				return nil
			},
			expFn: func(t *testing.T, err error) {
				require.Nil(t, err)
			},
		},
		{
			name:              "no interceptor: context canceled",
			noServerIntercept: true,
			errFn: func() error {
				return context.Canceled
			},
			expFn: func(t *testing.T, err error) {
				require.False(t, errors.Is(err, context.Canceled))
				require.Equal(t, "rpc error: code = Canceled desc = context canceled", err.Error())
			},
		},
		{
			name:              "no interceptor: context deadline exceeded",
			noServerIntercept: true,
			errFn: func() error {
				return context.DeadlineExceeded
			},
			expFn: func(t *testing.T, err error) {
				require.False(t, errors.Is(err, context.DeadlineExceeded))
				require.Equal(t, "rpc error: code = DeadlineExceeded desc = context deadline exceeded", err.Error())
			},
		},
		{
			name:              "no interceptor: fudge error",
			noServerIntercept: true,
			errFn: func() error {
				return errors.New("such test")
			},
			expFn: func(t *testing.T, err error) {
				require.Equal(t, "rpc error: code = Unknown desc = such test", err.Error())
			},
		},
		{
			name:              "no interceptor: no server",
			noServer:          true,
			noServerIntercept: true,
			expFn: func(t *testing.T, err error) {
				require.True(t, strings.Contains(err.Error(), `rpc error: code = Unavailable desc = connection error: desc = `))
			},
		},
		{
			name: "with interceptor: nil",
			errFn: func() error {
				return nil
			},
			expFn: func(t *testing.T, err error) {
				require.Nil(t, err)
			},
		},
		{
			name: "with interceptor: context canceled",
			errFn: func() error {
				return context.Canceled
			},
			expFn: func(t *testing.T, err error) {
				require.ErrorIs(t, err, context.Canceled)
				require.Equal(t, "rpc error: context canceled", err.Error())
			},
		},
		{
			name: "with interceptor: context deadline exceeded",
			errFn: func() error {
				return context.DeadlineExceeded
			},
			expFn: func(t *testing.T, err error) {
				require.ErrorIs(t, err, context.DeadlineExceeded)
				require.Equal(t, "rpc error: context deadline exceeded", err.Error())
			},
		},
		{
			name: "with interceptor: fudge error",
			errFn: func() error {
				return errors.New("such test")
			},
			expFn: func(t *testing.T, err error) {
				require.Equal(t, "rpc error: such test", err.Error())
			},
		},
		{
			name: "with interceptor: extra hop",
			hops: 1,
			errFn: func() error {
				return errors.New("such test")
			},
			expFn: func(t *testing.T, err error) {
				require.Equal(t, "rpc error: rpc error: such test", err.Error())
			},
		},
		{
			name:     "with interceptor: no server",
			noServer: true,
			expFn: func(t *testing.T, err error) {
				require.True(t, strings.Contains(err.Error(), `rpc error: code = Unavailable desc = connection error: desc = `))
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := fmt.Sprintf("localhost:%d", rand.Intn(10000)+10000)

			var serverOpts []grpc.ServerOption
			var clientOpts []grpc.DialOption
			if !tt.noServerIntercept {
				serverOpts = append(serverOpts,
					grpc.UnaryInterceptor(grpcerrors.UnaryServerInterceptor))
			}
			if !tt.noClientIntercept {
				clientOpts = append(clientOpts,
					grpc.WithUnaryInterceptor(grpcerrors.UnaryClientInterceptor))
			}

			if !tt.noServer {
				svr, err := grpctest.NewServer(addr)
				require.Nil(t, err)
				svr.SetErrFn(tt.errFn)

				gsvr := grpc.NewServer(serverOpts...)
				pb.RegisterCandyStoreServer(gsvr, svr)

				lis, err := net.Listen("tcp", addr)
				require.Nil(t, err)

				go gsvr.Serve(lis)
				defer gsvr.Stop()
			}

			client, err := grpctest.NewClient(addr, clientOpts...)
			require.Nil(t, err)

			err = client.Buy(ctx, tt.hops)
			tt.expFn(t, err)
		})
	}
}
