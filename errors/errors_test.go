package errors

import (
	"fmt"
	"io"
	"regexp"
	"testing"

	"github.com/rossmacarthur/fudge"
	"github.com/stretchr/testify/require"
)

var digits = regexp.MustCompile(`(_\w+\.s)?:\d+`)

func TestNew(t *testing.T) {
	err := New("such test", fudge.KV("key", "value"), fudge.KV("this", "that"))
	s := digits.ReplaceAllString(fmt.Sprintf("%+v", err), ":XXX")
	require.Equal(t,
		"github.com/rossmacarthur/fudge/errors/errors_test.go:XXX: such test {key:value, this:that}\n"+
			"testing/testing.go:XXX\n"+
			"runtime/asm:XXX", s)
}

func TestError(t *testing.T) {
	err := New("such test")
	s := digits.ReplaceAllString(err.Error(), ":XXX")
	require.Equal(t,
		"github.com/rossmacarthur/fudge/errors/errors_test.go:XXX: such test\n"+
			"testing/testing.go:XXX\n"+
			"runtime/asm:XXX", s)
}

func TestString(t *testing.T) {
	err := New("such test")
	s := digits.ReplaceAllString(err.(*Error).String(), ":XXX")
	require.Equal(t,
		"github.com/rossmacarthur/fudge/errors/errors_test.go:XXX: such test\n"+
			"testing/testing.go:XXX\n"+
			"runtime/asm:XXX", s)
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name  string
		errFn func() error
		exp   string
	}{
		{
			name:  "nil",
			errFn: func() error { return nil },
		},
		{
			name:  "std",
			errFn: func() error { return io.EOF },
			exp: `github.com/rossmacarthur/fudge/errors/errors_test.go:XXX: very wrap: EOF
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX: and another
testing/testing.go:XXX
runtime/asm:XXX`,
		},
		{
			name:  "fudge",
			errFn: func() error { return New("such test") },
			exp: `github.com/rossmacarthur/fudge/errors/errors_test.go:XXX: such test
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX: very wrap
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX: and another
testing/testing.go:XXX
runtime/asm:XXX`,
		},
		{
			name:  "fudge with kvs",
			errFn: func() error { return New("such test", fudge.KV("key", "value"), fudge.KV("this", "that")) },
			exp: `github.com/rossmacarthur/fudge/errors/errors_test.go:XXX: such test
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX: very wrap
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX: and another
testing/testing.go:XXX
runtime/asm:XXX`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err1 := Wrap(tc.errFn(), "very wrap", fudge.KV("a", "b")) // wrap within the stack trace
			err2 := Wrap(err1, "and another", fudge.KV("c", "d"))     // wrap outside the stack trace

			if err2 == nil {
				require.Equal(t, tc.exp, "")
			} else {
				s := digits.ReplaceAllString(err2.Error(), ":XXX")
				require.Equal(t, tc.exp, s)
			}
		})
	}
}

var errTest = New("test error")

func TestIs(t *testing.T) {
	err := Wrap(errTest, "it happened")
	require.True(t, Is(err, errTest))
	require.True(t, Is(errTest, errTest))
	require.False(t, Is(errTest, New("test error")))
	require.False(t, Is(errTest, nil))
}
