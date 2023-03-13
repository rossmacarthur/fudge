package fudgepb

import (
	"context"
	"sort"

	stderrors "errors"

	"github.com/rossmacarthur/fudge/errors"
)

const (
	kindStd   int32 = 1
	kindFudge int32 = 2
)

const (
	codeContextCanceled         = "context.Canceled"
	codeContextDeadlineExceeded = "context.DeadlineExceeded"
)

func FromProto(pb *Error) error {
	if pb == nil {
		return nil
	}

	var err error
	var done bool

	for i := len(pb.Hops) - 1; i >= 0; i-- {
		hop := pb.Hops[i]
		err, done = errorFromHop(hop, err)
		if done {
			break
		}
	}

	return err
}

func errorFromHop(hop *Hop, cause error) (error, bool) {
	switch hop.Kind {

	case kindStd:
		switch hop.Code {
		case codeContextCanceled:
			return context.Canceled, true
		case codeContextDeadlineExceeded:
			return context.DeadlineExceeded, true
		}
		return stderrors.New(hop.Message), true

	case kindFudge:
		hop := errors.Error{
			Binary:  hop.Binary,
			Message: hop.Message,
			Code:    hop.Code,
			Cause:   cause, // NB: Each error wraps the previous hop
			Trace:   traceFromProto(hop.Trace),
		}
		return &hop, false

	default:
		return errors.New("invalid data"), true
	}
}

func traceFromProto(pb []*Frame) []errors.Frame {
	trace := make([]errors.Frame, 0, len(pb))
	for _, f := range pb {
		trace = append(trace, errors.Frame{
			File:      f.File,
			Function:  f.Function,
			Line:      int(f.Line),
			Message:   f.Message,
			KeyValues: keyValuesFromProto(f.KeyValues),
		})
	}
	return trace
}

func keyValuesFromProto(pb []*KeyValue) errors.KeyValues {
	m := make(errors.KeyValues, len(pb))
	for _, kv := range pb {
		m[kv.Key] = kv.Value
	}
	return m
}

func ToProto(err error) *Error {
	var hops []*Hop

	for {
		if err == nil {
			break
		}
		hop, done := errorToHop(err)
		hops = append(hops, hop)
		if done {
			break
		}
		err = errors.Unwrap(err)
	}

	return &Error{Hops: hops}
}

func errorToHop(err error) (*Hop, bool) {
	ferr, ok := err.(*errors.Error)
	if ok {
		return &Hop{
			Kind:    kindFudge,
			Binary:  ferr.Binary,
			Message: ferr.Message,
			Code:    ferr.Code,
			Trace:   traceToProto(ferr.Trace),
		}, false
	}

	// Not a Fudge error, see if it is a supported sentinel
	var code string
	if errors.Is(err, context.Canceled) {
		code = codeContextCanceled
	} else if errors.Is(err, context.DeadlineExceeded) {
		code = codeContextDeadlineExceeded
	}

	return &Hop{
		Kind:    kindStd,
		Message: err.Error(),
		Code:    code,
	}, true
}

func traceToProto(trace []errors.Frame) []*Frame {
	pb := make([]*Frame, 0, len(trace))
	for _, f := range trace {
		pb = append(pb, &Frame{
			File:      f.File,
			Function:  f.Function,
			Line:      int32(f.Line),
			Message:   f.Message,
			KeyValues: keyValuesToProto(f.KeyValues),
		})
	}
	return pb
}

func keyValuesToProto(m errors.KeyValues) []*KeyValue {
	if len(m) == 0 {
		return nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pb := make([]*KeyValue, 0, len(m))
	for _, k := range keys {
		pb = append(pb, &KeyValue{
			Key:   k,
			Value: m[k],
		})
	}

	return pb
}
