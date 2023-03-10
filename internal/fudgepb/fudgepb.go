package fudgepb

import (
	"sort"

	"github.com/rossmacarthur/fudge/errors"
)

func FromProto(pb *Error) error {
	if pb == nil {
		return nil
	}

	var err error

	for i := len(pb.Hops) - 1; i >= 0; i-- {
		hop := pb.Hops[i]
		err = &errors.Error{
			Binary:   hop.Binary,
			Message:  hop.Message,
			Code:     hop.Code,
			Original: err, // NB: Each error wraps the previous hop
			Trace:    traceFromProto(hop.Trace),
		}
	}

	return err
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

		ferr, ok := err.(*errors.Error)
		if ok {
			hops = append(hops, &Hop{
				Binary:  ferr.Binary,
				Message: ferr.Message,
				Code:    ferr.Code,
				Trace:   traceToProto(ferr.Trace),
			})
			err = ferr.Unwrap()
		} else {
			// Not a Fudge error, so, theres nothing more we can do
			hops = append(hops, &Hop{
				Message: err.Error(),
			})
			break
		}
	}

	return &Error{Hops: hops}
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
