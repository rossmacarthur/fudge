package fudgepb_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/rossmacarthur/fudge"
	"github.com/rossmacarthur/fudge/errors"
	"github.com/rossmacarthur/fudge/internal/fudgepb"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

var errSentinel = errors.NewSentinel("such test", "TEST1234")

func TestFromProto(t *testing.T) {
	tests := []struct {
		name string
		err  *fudgepb.Error
	}{
		{
			name: "empty trace",
			err: &fudgepb.Error{
				Hops: []*fudgepb.Hop{
					{
						Binary:  "fudgepb.test",
						Message: "such test",
					},
				},
			},
		},
		{
			name: "one hop",
			err: &fudgepb.Error{
				Hops: []*fudgepb.Hop{
					{
						Binary: "fudgepb.test",
						Trace: []*fudgepb.Frame{
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
			name: "two hops",
			err: &fudgepb.Error{
				Hops: []*fudgepb.Hop{
					{
						Binary: "fudgepb.test",
						Trace: []*fudgepb.Frame{
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
						Binary: "fudgepb.test",
						Trace: []*fudgepb.Frame{
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
	}

	g := goldie.New(t, goldie.WithTestNameForDir(true))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fudgepb.FromProto(tt.err)
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
			got := fudgepb.ToProto(tt.errFn())
			bytes, err := protojson.MarshalOptions{Multiline: true}.Marshal(got)
			require.Nil(t, err)
			g.Assert(t, tt.name, bytes)
		})
	}
}

func TestRoundtrip(t *testing.T) {
	err := errors.Wrap(errSentinel, "very wrap", fudge.KV("foo", "bar"))
	got := fudgepb.FromProto(fudgepb.ToProto(err))
	require.True(t, errors.Is(got, err))
	require.True(t, errors.Is(got, errSentinel))

	err = errors.New("such test", fudge.KV("foo", "bar"))
	got = fudgepb.FromProto(fudgepb.ToProto(err))
	require.False(t, errors.Is(got, err))
	require.False(t, errors.Is(got, errSentinel))
}
