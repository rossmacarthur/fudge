package grpc

import (
	"context"

	"github.com/rossmacarthur/fudge/errors"
	"github.com/rossmacarthur/fudge/internal/fudgepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// interceptClient tries converting the error to a gRPC status and if it can
// then it extracts any Fudge information out of the details. Otherwise the
// error is simply wrapped to add a stack trace.
func interceptClient(err error) error {
	if err == nil {
		return nil
	}
	s, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error
		return errors.Wrap(err, "")
	}
	return FromStatus(s)
}

// interceptServer converts the error into an error that implements GRPCStatus.
// Any Fudge error information is encoded in the gRPC status details.
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
