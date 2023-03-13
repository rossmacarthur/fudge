package fudgepb

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/rossmacarthur/fudge"
	"github.com/rossmacarthur/fudge/errors"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

var errSentinel = errors.NewSentinel("such test", "TEST1234")

func TestFromProto(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
	}{
		{
			name: "empty trace",
			err: &Error{
				Hops: []*Hop{
					{
						Kind:    kindFudge,
						Binary:  "fudgepb.test",
						Message: "such test",
					},
				},
			},
		},
		{
			name: "one fudge hop",
			err: &Error{
				Hops: []*Hop{
					{
						Kind:   kindFudge,
						Binary: "fudgepb.test",
						Trace: []*Frame{
							{
								File:     "github.com/rossmacarthur/fudge/internal/fudgepb/fudgepb_test.go",
								Function: "TestFromProto",
								Line:     24,
								Message:  "such test",
							},
							{
								File:     "testing/testing.go",
								Function: "tRunner",
								Line:     1576,
							},
							{
								File:     "runtime/asm_arm64.s",
								Function: "goexit",
								Line:     1172,
							},
						},
					},
				},
			},
		},
		{
			name: "two fudge hops",
			err: &Error{
				Hops: []*Hop{
					{
						Kind:   kindFudge,
						Binary: "fudgepb.test",
						Trace: []*Frame{
							{
								File:     "github.com/rossmacarthur/fudge/internal/fudgepb/fudgepb_test.go",
								Function: "TestFromProto",
								Line:     52,
								Message:  "this hop",
							},
							{
								File:     "testing/testing.go",
								Function: "tRunner",
								Line:     1576,
							},
							{
								File:     "runtime/asm_arm64.s",
								Function: "goexit",
								Line:     1172,
							},
						},
					},
					{
						Kind:   kindFudge,
						Binary: "fudgepb.test",
						Trace: []*Frame{
							{
								File:     "github.com/rossmacarthur/fudge/internal/fudgepb/fudgepb_test.go",
								Function: "TestFromProto",
								Line:     52,
								Message:  "such test",
							},
							{
								File:     "github.com/rossmacarthur/fudge/internal/fudgepb/fudgepb_test.go",
								Function: "TestFromProto",
								Line:     53,
								Message:  "very wrap",
							},
							{
								File:     "testing/testing.go",
								Function: "tRunner",
								Line:     1576,
							},
							{
								File:     "runtime/asm_arm64.s",
								Function: "goexit",
								Line:     1172,
							},
						},
					},
				},
			},
		},
		{
			name: "one std hop",
			err: &Error{
				Hops: []*Hop{
					{
						Kind:    kindStd,
						Message: "context canceled",
						Code:    codeContextCanceled,
					},
				},
			},
		},
		{
			name: "one fudge hop, one std hop",
			err: &Error{
				Hops: []*Hop{
					{
						Kind:   kindFudge,
						Binary: "fudgepb.test",
						Trace: []*Frame{
							{
								File:     "github.com/rossmacarthur/fudge/internal/fudgepb/fudgepb_test.go",
								Function: "TestFromProto",
								Line:     53,
								Message:  "very wrap",
							},
							{
								File:     "testing/testing.go",
								Function: "tRunner",
								Line:     1576,
							},
							{
								File:     "runtime/asm_arm64.s",
								Function: "goexit",
								Line:     1172,
							},
						},
					},

					{
						Kind:    kindStd,
						Message: "context canceled",
						Code:    codeContextCanceled,
					},
				},
			},
		},
	}

	g := goldie.New(t, goldie.WithTestNameForDir(true))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FromProto(tt.err)
			g.Assert(t, tt.name, []byte(fmt.Sprintf("%+v", err)))
		})
	}
}

func TestToProto(t *testing.T) {
	tests := []struct {
		name  string
		errFn func() error
	}{
		{
			name:  "nil",
			errFn: func() error { return nil },
		},
		{
			name:  "std",
			errFn: func() error { return io.EOF },
		},
		{
			name:  "std sentinel",
			errFn: func() error { return context.Canceled },
		},
		{
			name:  "std sentinel wrapped",
			errFn: func() error { return errors.Wrap(context.Canceled, "very wrap") },
		},
		{
			name:  "fudge",
			errFn: func() error { return errors.New("such test") },
		},
		{
			name: "fudge wrapped",
			errFn: func() error {
				err := errors.New("such test")
				return errors.Wrap(err, "very wrap")
			},
		},
		{
			name: "fudge with kvs",
			errFn: func() error {
				err := errors.New("such test", fudge.KV("hello", "world"))
				return errors.Wrap(err, "very wrap", fudge.KV("foo", "bar"))
			},
		},
		{
			name: "fudge sentinel",
			errFn: func() error {
				return errors.Wrap(errSentinel, "very wrap")
			},
		},
		{
			name: "with hops",
			errFn: func() error {
				err := errors.New("this hop")
				ferr := new(errors.Error)
				require.True(t, errors.As(err, &ferr))
				ferr.Cause = errors.Wrap(errSentinel, "very wrap")
				return err
			},
		},
	}

	g := goldie.New(t, goldie.WithTestNameForDir(true))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToProto(tt.errFn())
			bytes, err := protojson.MarshalOptions{Multiline: true}.Marshal(got)
			require.Nil(t, err)
			g.Assert(t, tt.name, bytes)
		})
	}
}

func TestRoundtrip(t *testing.T) {
	err := errors.Wrap(errSentinel, "very wrap", fudge.KV("foo", "bar"))
	got := FromProto(ToProto(err))
	require.True(t, errors.Is(got, err))
	require.True(t, errors.Is(got, errSentinel))

	err = errors.New("such test", fudge.KV("foo", "bar"))
	got = FromProto(ToProto(err))
	require.False(t, errors.Is(got, err))
	require.False(t, errors.Is(got, errSentinel))

	err = errors.Wrap(context.Canceled, "very wrap", fudge.KV("foo", "bar"))
	got = FromProto(ToProto(err))
	require.True(t, errors.Is(got, context.Canceled))
	require.False(t, errors.Is(got, context.DeadlineExceeded))
	require.False(t, errors.Is(got, err))
	require.False(t, errors.Is(got, errSentinel))
}
