package grpcerrors

import (
	"context"

	"github.com/rossmacarthur/fudge/errors"
	"github.com/rossmacarthur/fudge/internal/fudgepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func interceptClient(err error) error {
	s, _ := status.FromError(err)
	return FromStatus(s)
}

func interceptServer(err error) error {
	if err == nil {
		return nil
	}
	return &grpcError{err: err}
}

// grpcError wraps an error and implements the GRPCStatus interface.
type grpcError struct {
	err error
}

// Error implements the error interface
func (e *grpcError) Error() string {
	return e.err.Error()
}

// GRPCStatus implements the interface necessary to convert an error into a
// gRPC status. Fudge errors are converted into a protobuf representation and
// stored in the details.
func (e *grpcError) GRPCStatus() *status.Status {
	code := codes.Unknown
	if errors.Is(e.err, context.Canceled) {
		code = codes.Canceled
	} else if errors.Is(e.err, context.DeadlineExceeded) {
		code = codes.DeadlineExceeded
	}

	s := status.New(code, e.err.Error())
	sw, err := s.WithDetails(fudgepb.ToProto(e.err))
	if err != nil {
		// TODO: Log in this case?
		return s
	}

	return sw
}

// FromStatus converts a gRPC status into an error by extracting any Fudge
// information from the details.
func FromStatus(s *status.Status) error {
	if s.Code() == codes.OK {
		return nil
	}

	for _, d := range s.Details() {
		pb, ok := d.(*fudgepb.Error)
		if !ok {
			continue
		}
		// NB: Don't wrap because we want to start a new hop.
		return errors.NewWithCause("rpc error", fudgepb.FromProto(pb))
	}

	return errors.Wrap(s.Err(), "")
}
