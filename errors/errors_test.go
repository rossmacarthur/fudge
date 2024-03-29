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

var sentinelTest = Sentinel("test error", "TEST1234")

type stringError struct {
	msg string
}

func (e *stringError) Error() string {
	return e.msg
}

func TestNew(t *testing.T) {
	err := New("such test",
		fudge.KV("key", "value"),
		fudge.KV("this", "that"),
		fudge.MKV{"other": "thing"})
	s := digits.ReplaceAllString(fmt.Sprintf("%#v", err), ":XXX")
	require.Equal(t, `such test {key:value, other:thing, this:that}
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX TestNew
testing/testing.go:XXX tRunner
runtime/asm:XXX goexit`, s)
}

func TestNewSentinel(t *testing.T) {
	err := Sentinel("such test", "TEST1234")
	s := digits.ReplaceAllString(fmt.Sprintf("%#v", err), ":XXX")
	require.Equal(t, "such test (TEST1234)", s)
}

func TestErrorAndString(t *testing.T) {
	tests := []struct {
		name string
		err  error
		exp  string
	}{
		{
			name: "basic",
			err:  New("such test"),
			exp:  "such test",
		},
		{
			name: "wrapped",
			err:  Wrap(New("such test"), "very wrap"),
			exp:  "very wrap: such test",
		},
		{
			name: "wrapped twice",
			err:  Wrap(Wrap(New("such test"), "very wrap"), "and another"),
			exp:  "and another: very wrap: such test",
		},
		{
			name: "wrapped non-Fudge",
			err:  Wrap(io.ErrClosedPipe, "very wrap"),
			exp:  "very wrap: io: read/write on closed pipe",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.exp, tc.err.Error())
		})
	}
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
			name:  "other",
			errFn: func() error { return &stringError{msg: "such test"} },
			exp: `and another: very wrap: such test
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX TestWrap.func7
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX TestWrap.func7
testing/testing.go:XXX tRunner
runtime/asm:XXX goexit`,
		},
		{
			name:  "std",
			errFn: func() error { return io.EOF },
			exp: `and another: very wrap: EOF
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX TestWrap.func7
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX TestWrap.func7
testing/testing.go:XXX tRunner
runtime/asm:XXX goexit`,
		},
		{
			name:  "fudge",
			errFn: func() error { return New("such test") },
			exp: `and another: very wrap: such test
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX TestWrap.func4
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX TestWrap.func7
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX TestWrap.func7
testing/testing.go:XXX tRunner
runtime/asm:XXX goexit`,
		},
		{
			name:  "fudge with kvs",
			errFn: func() error { return New("such test", fudge.KV("key", "value"), fudge.KV("this", "that")) },
			exp: `and another: very wrap: such test
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX TestWrap.func5
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX TestWrap.func7
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX TestWrap.func7
testing/testing.go:XXX tRunner
runtime/asm:XXX goexit`,
		},
		{
			name:  "fudge sentinel",
			errFn: func() error { return Sentinel("such test", "123456") },
			exp: `and another: very wrap: such test (123456)
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX TestWrap.func7
github.com/rossmacarthur/fudge/errors/errors_test.go:XXX TestWrap.func7
testing/testing.go:XXX tRunner
runtime/asm:XXX goexit`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err1 := Wrap(tc.errFn(), "very wrap", fudge.KV("a", "b")) // wrap within the stack trace
			err2 := Wrap(err1, "and another", fudge.KV("c", "d"))     // wrap outside the stack trace

			if err2 == nil {
				require.Equal(t, tc.exp, "")
			} else {
				s := digits.ReplaceAllString(fmt.Sprintf("%+v", err2), ":XXX")
				require.Equal(t, tc.exp, s)
			}
		})
	}
}

func TestIs(t *testing.T) {
	// local
	errTest := New("test error")
	require.True(t, Is(errTest, errTest))
	require.False(t, Is(errTest, New("test error")))
	require.False(t, Is(errTest, nil))
	require.False(t, Is(errTest, nil))

	// sentinel
	require.True(t, Is(sentinelTest, sentinelTest))
	require.False(t, Is(sentinelTest, New("test error")))
	require.False(t, Is(sentinelTest, nil))
	require.False(t, Is(sentinelTest, nil))

	// wrapped
	err := Wrap(sentinelTest, "very wrap")
	require.True(t, Is(err, sentinelTest))
	require.False(t, Is(err, New("test error")))
	require.False(t, Is(err, nil))

	// wrapped twice
	err = Wrap(Wrap(sentinelTest, "very wrap"), "it happened")
	require.True(t, Is(err, sentinelTest))
	require.False(t, Is(sentinelTest, New("test error")))
	require.False(t, Is(sentinelTest, nil))

	// wrapped non-Fudge
	err = Wrap(io.EOF, "very wrap")
	require.True(t, Is(err, io.EOF))
	require.False(t, Is(err, sentinelTest))
	require.False(t, Is(sentinelTest, New("EOF")))
	require.False(t, Is(err, nil))
}

func TestAs(t *testing.T) {
	te := new(Error)
	ts := new(stringError)

	// local
	errTest := New("test error")
	require.True(t, As(errTest, &te))
	require.False(t, As(errTest, &ts))

	// sentinel
	require.True(t, As(sentinelTest, &te))
	require.False(t, As(sentinelTest, &ts))

	// wrapped
	err := Wrap(sentinelTest, "very wrap")
	require.True(t, As(err, &te))
	require.False(t, As(err, &ts))

	// wrapped twice
	err = Wrap(Wrap(sentinelTest, "very wrap"), "it happened")
	require.True(t, As(err, &te))
	require.False(t, As(err, &ts))

	// wrapped non-Fudge
	var serr error = &stringError{msg: "test error"}
	require.True(t, As(serr, &ts))
	require.False(t, As(serr, &te))
}
