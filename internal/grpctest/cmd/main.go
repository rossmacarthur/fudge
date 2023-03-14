package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/rossmacarthur/fudge/errors"
	grpcerrors "github.com/rossmacarthur/fudge/grpc/errors"
	"github.com/rossmacarthur/fudge/internal/grpctest"
	"github.com/rossmacarthur/fudge/internal/grpctest/pb"
	"google.golang.org/grpc"
)

var (
	asClient = flag.Int64("client", -1, "run as client")
	asServer = flag.Bool("server", false, "run as server")
)

func main() {
	flag.Parse()

	if *asServer == (*asClient != -1) {
		log.Fatal("must specify either -client or -server")
	}

	addr := "localhost:8000"

	if *asServer {
		svr, err := grpctest.NewServer(addr)
		if err != nil {
			log.Fatal(err)
		}

		svr.SetErrFn(func() error {
			return errors.New("such test")
		})

		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatal(err)
		}

		gsvr := grpc.NewServer(
			grpc.UnaryInterceptor(grpcerrors.UnaryServerInterceptor))
		pb.RegisterCandyStoreServer(gsvr, svr)

		fmt.Printf("Serving gRPC on %s...\n", addr)
		gsvr.Serve(lis)

	} else {
		ctx := context.Background()
		cl, err := grpctest.NewClient(addr,
			grpc.WithUnaryInterceptor(grpcerrors.UnaryClientInterceptor))
		if err != nil {
			log.Fatal(err)
		}
		err = cl.Buy(ctx, *asClient)
		for err != nil {
			fmt.Printf("\n%+v\n", err)
			err = errors.Unwrap(err)
		}
	}
}
