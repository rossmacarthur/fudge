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

var errSentinel = errors.NewSentinel("such test", "ERR_12345")

func TestInterceptors(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string

		// noServer doesn't run the gRPC server
		noServer bool

		// noServerIntercept doesn't add the server gRPC interceptors
		noServerIntercept bool

		// noClientIntercept doesn't add the client gRPC interceptors
		noClientIntercept bool

		// errFn generates the error on the server
		errFn func() error

		// expFn asserts any conditions this test case requires
		expFn func(t *testing.T, client *grpctest.Client)
	}{
		{
			name:              "unary: no interceptor: nil",
			noServerIntercept: true,
			noClientIntercept: true,
			errFn: func() error {
				return nil
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.Buy(ctx, 0)
				require.Nil(t, err)
			},
		},
		{
			name:              "unary: no interceptor: context canceled",
			noServerIntercept: true,
			noClientIntercept: true,
			errFn: func() error {
				return context.Canceled
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.Buy(ctx, 0)
				require.False(t, isFudge(err))
				require.False(t, errors.Is(err, context.Canceled))
				require.Equal(t, "rpc error: code = Canceled desc = context canceled", err.Error())
			},
		},
		{
			name:              "unary: no interceptor: context deadline exceeded",
			noServerIntercept: true,
			noClientIntercept: true,
			errFn: func() error {
				return context.DeadlineExceeded
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.Buy(ctx, 0)
				require.False(t, isFudge(err))
				require.False(t, errors.Is(err, context.DeadlineExceeded))
				require.Equal(t, "rpc error: code = DeadlineExceeded desc = context deadline exceeded", err.Error())
			},
		},
		{
			name:              "unary: no interceptor: fudge error",
			noServerIntercept: true,
			noClientIntercept: true,
			errFn: func() error {
				return errors.New("such test")
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.Buy(ctx, 0)
				require.False(t, isFudge(err))
				require.Equal(t, "rpc error: code = Unknown desc = such test", err.Error())
			},
		},
		{
			name:              "unary: no interceptor: no server",
			noServer:          true,
			noServerIntercept: true,
			noClientIntercept: true,
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.Buy(ctx, 0)
				require.False(t, isFudge(err))
				require.True(t, strings.Contains(err.Error(), `rpc error: code = Unavailable desc = connection error: desc = `))
			},
		},
		{
			name: "unary: with interceptor: nil",
			errFn: func() error {
				return nil
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.Buy(ctx, 0)
				require.Nil(t, err)
			},
		},
		{
			name: "unary: with interceptor: context canceled",
			errFn: func() error {
				return context.Canceled
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.Buy(ctx, 0)
				require.ErrorIs(t, err, context.Canceled)
				require.Equal(t, "rpc error: context canceled", err.Error())
			},
		},
		{
			name: "unary: with interceptor: context deadline exceeded",
			errFn: func() error {
				return context.DeadlineExceeded
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.Buy(ctx, 0)
				require.ErrorIs(t, err, context.DeadlineExceeded)
				require.Equal(t, "rpc error: context deadline exceeded", err.Error())
			},
		},
		{
			name: "unary: with interceptor: fudge error",
			errFn: func() error {
				return errors.New("such test")
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.Buy(ctx, 0)
				require.Equal(t, "rpc error: such test", err.Error())
			},
		},
		{
			name: "unary: with interceptor: fudge sentinel error",
			errFn: func() error {
				return errors.Wrap(errSentinel, "")
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.Buy(ctx, 0)
				require.ErrorIs(t, err, errSentinel)
				require.Equal(t, "rpc error: such test (ERR_12345)", err.Error())
			},
		},
		{
			name: "unary: with interceptor: extra hop",
			errFn: func() error {
				return errors.New("such test")
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.Buy(ctx, 1)
				require.Equal(t, "rpc error: rpc error: such test", err.Error())
			},
		},
		{
			name:     "unary: with interceptor: no server",
			noServer: true,
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.Buy(ctx, 0)
				require.True(t, isFudge(err))
				require.True(t, strings.Contains(err.Error(), `rpc error: code = Unavailable desc = connection error: desc = `))
			},
		},
		{
			name: "stream from: nil",
			errFn: func() error {
				return nil
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				candy, err := client.StreamCandyFrom(ctx)
				require.Nil(t, err)
				require.Equal(t, []string{"chocolate", "gummy", "lollipop"}, candy)
			},
		},
		{
			name: "stream from: context canceled",
			errFn: func() error {
				return context.Canceled
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				candy, err := client.StreamCandyFrom(ctx)
				require.True(t, isFudge(err))
				require.ErrorIs(t, err, context.Canceled)
				require.Equal(t, "rpc error: context canceled", err.Error())
				require.Nil(t, candy)
			},
		},
		{
			name: "stream from: context deadline exceeded",
			errFn: func() error {
				return context.DeadlineExceeded
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				candy, err := client.StreamCandyFrom(ctx)
				require.True(t, isFudge(err))
				require.ErrorIs(t, err, context.DeadlineExceeded)
				require.Equal(t, "rpc error: context deadline exceeded", err.Error())
				require.Nil(t, candy)
			},
		},
		{
			name: "stream from: fudge error",
			errFn: func() error {
				return errors.New("such test")
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				candy, err := client.StreamCandyFrom(ctx)
				require.True(t, isFudge(err))
				require.Equal(t, "rpc error: such test", err.Error())
				require.Nil(t, candy)
			},
		},
		{
			name: "stream from: fudge sentinel error",
			errFn: func() error {
				return errors.Wrap(errSentinel, "")
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				candy, err := client.StreamCandyFrom(ctx)
				require.True(t, isFudge(err))
				require.ErrorIs(t, err, errSentinel)
				require.Equal(t, "rpc error: such test (ERR_12345)", err.Error())
				require.Nil(t, candy)
			},
		},
		{
			name: "stream to: nil",
			errFn: func() error {
				return nil
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.StreamCandyTo(ctx, []string{"whispers"})
				require.Nil(t, err)
			},
		},
		{
			name: "stream to: context canceled",
			errFn: func() error {
				return context.Canceled
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.StreamCandyTo(ctx, []string{"whispers"})
				require.True(t, isFudge(err))
				require.ErrorIs(t, err, context.Canceled)
				require.Equal(t, "rpc error: context canceled", err.Error())
			},
		},
		{
			name: "stream to: context deadline exceeded",
			errFn: func() error {
				return context.DeadlineExceeded
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.StreamCandyTo(ctx, []string{"whispers"})
				require.True(t, isFudge(err))
				require.ErrorIs(t, err, context.DeadlineExceeded)
				require.Equal(t, "rpc error: context deadline exceeded", err.Error())
			},
		},
		{
			name: "stream to: fudge error",
			errFn: func() error {
				return errors.New("such test")
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.StreamCandyTo(ctx, []string{"whispers"})
				require.True(t, isFudge(err))
				require.Equal(t, "rpc error: such test", err.Error())
			},
		},
		{
			name: "stream to: fudge sentinel error",
			errFn: func() error {
				return errors.Wrap(errSentinel, "")
			},
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.StreamCandyTo(ctx, []string{"whispers"})
				require.True(t, isFudge(err))
				require.ErrorIs(t, err, errSentinel)
				require.Equal(t, "rpc error: such test (ERR_12345)", err.Error())
			},
		},
		{
			name: "stream to: in stock",
			expFn: func(t *testing.T, client *grpctest.Client) {
				err := client.StreamCandyTo(ctx, []string{"chocolate"})
				require.True(t, isFudge(err))
				require.Equal(t, "rpc error: already in stock", err.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := fmt.Sprintf("localhost:%d", rand.Intn(10000)+10000)

			var serverOpts []grpc.ServerOption
			var clientOpts []grpc.DialOption
			if !tt.noServerIntercept {
				serverOpts = append(serverOpts,
					grpc.UnaryInterceptor(grpcerrors.UnaryServerInterceptor),
					grpc.StreamInterceptor(grpcerrors.StreamServerInterceptor))
			}
			if !tt.noClientIntercept {
				clientOpts = append(clientOpts,
					grpc.WithUnaryInterceptor(grpcerrors.UnaryClientInterceptor),
					grpc.WithStreamInterceptor(grpcerrors.StreamClientInterceptor))
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
			tt.expFn(t, client)
		})
	}
}

func isFudge(err error) bool {
	_, ok := err.(*errors.Error)
	return ok
}
