# fudge

Oh Fudge! A straight-forward error and logging library for Go.

## Installation

Install using the following

```
go get github.com/rossmacarthur/fudge
```

## fudge/errors

fudge/errors provides a simple way to add contextual messages, structured key
values and stack traces to errors. You can import the package using the
following.

```go
import "github.com/rossmacarthur/fudge/errors"
```

### Construction

Inline errors can be produced in a familiar way. A stack trace will be attached
to the error.

```go
errors.New("failed to shave yak")
```

Additional key value pairs can be provided which can be inspected later.

```go
errors.New("failed to shave yak", fudge.KV("yak_id", yakID))
```

Multiple key value pairs are also possible.

```go
errors.New("failed to shave yak", fudge.MKV{"yak_id": yakID, "hair_len": hairLen})
```

Sentinel errors are defined in the global scope and do not get a stack trace
attached until they are wrapped with Wrap.

```go
var ErrShavingFailed = errors.New("failed to shave yak")
```

### Annotation

Existing errors can be wrapped with contextual messages and key value pairs. You
can wrap both Fudge errors and non-Fudge errors. The only difference is that the
traceback will start at the `errors.Wrap` call for non-Fudge errors.

```go
func locateRazor() error {
    return errors.New("failed to locate razor", fudge.KV("hair_len", hairLen))
}

err := locateRazor()
if err != nil {
    return errors.Wrap(err, "failed to shave yak", fudge.KV("yak_id", yakID))
}
```

Sentinel errors must be wrapped when used so that they get a stack trace. Errors
can be compared to sentinel errors using `errors.Is`.

```go
var ErrRazorNotFound = errors.New("razor not found")

func locateRazor() error {
    return errors.Wrap(ErrRazorNotFound, "", fudge.KV("hair_len", hairLen))
}

err := locateRazor()
if errors.Is(err, ErrRazorNotFound) {
    // use backup razor
} else if err != nil {
    return errors.Wrap(err, "failed to shave yak", fudge.KV("yak_id", yakID))
}
```

### Formatting

For example given the following.

```go
var ErrRazorNotFound = errors.New("razor not found")

func locateRazor() error {
    return errors.Wrap(ErrRazorNotFound, "", fudge.KV("hair_len", hairLen))
}

err := locateRazor()
if err != nil {
    return errors.Wrap(err, "failed to shave yak", fudge.KV("yak_id", yakID))
}
```

By default `.Error()` output will return a colon (`: `) separated list of
contextual messages like this:

```
failed to shave yak: razor not found
```

If formatted using `fmt.Sprintf("%+v", err)` the stack trace will be shown with
any added contextual messages.

```text
failed to shave yak: razor not found
example/main.go:20 locateRazor
example/main.go:24 example
example/main.go:26 example
example/main.go:13 main
runtime/proc.go:250 main
runtime/asm_arm64.s:1172 goexit
```

If formatted using `fmt.Sprintf("%#v", err)` then the key value pairs will also
be added.

```text
failed to shave yak: razor not found {hair_len:7, yak_id:1337}
example/main.go:20 locateRazor
example/main.go:24 example
example/main.go:26 example
example/main.go:13 main
runtime/proc.go:250 main
runtime/asm_arm64.s:1172 goexit
```

Custom formatting is possible. For example:

```go
ferr := new(errors.Error)
if ok := errors.As(err, &ferr); ok {
    fmt.Printf("Code:\n  %s\nTraceback:\n", ferr.Code)
    for _, frame := range ferr.Trace {
        fmt.Printf("%s:%d\n", frame.File, frame.Line)
    }
}
```

## License

This project is distributed under the terms of both the MIT license and the
Apache License (Version 2.0).

See [LICENSE-APACHE](LICENSE-APACHE) and [LICENSE-MIT](LICENSE-MIT) for details.
