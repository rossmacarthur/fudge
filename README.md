# fudge

[![Go Reference](https://pkg.go.dev/badge/github.com/rossmacarthur/fudge.svg)](https://pkg.go.dev/github.com/rossmacarthur/fudge)
[![Build Status](https://github.com/rossmacarthur/fudge/workflows/build/badge.svg)](https://github.com/rossmacarthur/fudge/actions/workflows/build.yaml)

Oh Fudge! A straight-forward error library for Go.

![example](https://github.com/rossmacarthur/fudge/assets/17109887/dcbc97f1-07b9-4ee5-8211-380b09766d14)

## Features

- Implements idiomatic standard library functions
  - Unwrap wrapped errors using `errors.Unwrap`
  - Compare errors with `errors.Is`
- Contextual messages
- Structured key value pairs
- Stack traces
- Custom formatting
- gRPC support
- Multi error support (planned)

## Getting started

Install using the following

```
go get github.com/rossmacarthur/fudge
```

And import this package instead of the standard library `errors`

```go
import "github.com/rossmacarthur/fudge/errors"
```

Construct errors in the usual way

```go
errors.New("razor not found")
```

Or wrap a standard library or Fudge sentinel

```go
errors.Wrap(io.EOF, "failed to shave yak")
```

The error will be constructed with a stack trace attached. Now the error can be
passed up the call stack like normal.

```go
if err != nil {
    return err
}
```

Optionally, existing errors can be wrapped with additional context and key
value pairs.

```go
if err != nil {
    return errors.Wrap(err, "failed to shave yak", fudge.KV("yak_id", yakID))
}
```

Formatting this error with `%+v` results in the following

```go
fmt.Printf("%+v\n", err)
```
```text
failed to shave yak: razor not found
example/razor.go:20 locateRazor
example/razor.go:26 example
example/main.go:13 main
runtime/proc.go:250 main
runtime/asm_arm64.s:1172 goexit
```

## Construction

Error construction is straight-forward. Fudge considers messages formatted with
contextual information to be an antipattern so no `Newf` function is provided.
Instead key value pairs should be used.

### Inline errors

- Ordinary inline errors

  ```go
  errors.New("failed to shave yak")
  ```

- Inline errors with a key value pair

  ```go
  errors.New("failed to shave yak", fudge.KV("yak_id", yakID))
  ```

- Inline errors with multiple key value pairs

  ```go
  errors.New("failed to shave yak", fudge.MKV{"yak_id": yakID, "hair_len": hairLen})
  ```

- Wrap an existing Fudge sentinel, this adds a stack trace

  ```go
  errors.Wrap(ErrRazorNotFound, "failed to shave yak")
  ```

- Wrap an existing non-Fudge error, this wraps it with a Fudge error

  ```go
  _, err := os.ReadFile(path)
  if err != nil {
      return errors.Wrap(err, "failed to read file")
  }
  ```

- Wrap an existing non-Fudge sentinel, this wraps it with a Fudge error

  ```go
  errors.Wrap(io.EOF, "failed to read file")
  ```

### Sentinel errors

Sentinel errors are typically defined in the global scope and do not get a stack
trace attached until they are wrapped with `Wrap`. A code is required in order
for the error to be passed across gRPC in such a way that `errors.Is` checks
still work. If you don't need this behaviour then you can just define sentinels
using `errors.New`.


```go
// If you need gRPC support
var ErrShavingFailed = errors.Sentinel("failed to shave yak", "ERR_0a8cba3dfa944ecb")

// Otherwise this is fine
var ErrShavingFailed = errors.New("failed to shave yak")
```

💡 The [`fudge`](#command) command can automatically generate sentinel error
codes for you.

## Comparisons

Any error can be compared against sentinels using `errors.Is`  no matter how
many times they were wrapped.

```go
var ErrRazorNotFound = errors.New("razor not found")

if errors.Is(err, io.EOF) {
    // ...
} else if errors.Is(err, ErrRazorNotFound) {
    // ...
}
```

However, this doesn't work when errors are passed over the wire using gRPC. In
order for that to work the sentinel error must be defined using
`errors.Sentinel` which assigns a unique code. This code is then passed over the
wire.

```go
var ErrRazorNotFound = errors.Sentinel("razor not found", "ERR_0a8cba3dfa944ecb")
```

## Formatting

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

## gRPC interceptors

The `errors/grpc` package provides gRPC interceptors that can be used to
serialize and deserialize errors over the wire.

On the server add the following interceptors.

```go
import (
    errorsgrpc "github.com/rossmacarthur/fudge/errors/grpc"
)

grpc.NewServer(
    grpc.UnaryInterceptor(errorsgrpc.UnaryServerInterceptor),
    grpc.StreamInterceptor(errorsgrpc.StreamServerInterceptor))
```

On the client add the following interceptors.
```go
import (
    errorsgrpc "github.com/rossmacarthur/fudge/errors/grpc"
)

grpc.DialContext(ctx, addr,
    grpc.WithUnaryInterceptor(errorsgrpc.UnaryClientInterceptor),
    grpc.WithStreamInterceptor(errorsgrpc.StreamClientInterceptor))
```

## Command

The `fudge` command is provided to automatically generate error codes for
sentinel errors.

```
go install github.com/rossmacarthur/fudge/cmd/fudge
```

The command recursively rewrites Go files in a directory to add codes. E.g.
given the following code.

```go
var ErrShavingFailed = errors.Sentinel("failed to shave yak", "")
```

It would be automatically updated to something like this.

```go
var ErrShavingFailed = errors.Sentinel("failed to shave yak", "ERR_0a8cba3dfa944ecb")
```

## Acknowledgements

Inspired by [github.com/luno/jettison](https://github.com/luno/jettison).

## License

This project is distributed under the terms of both the MIT license and the
Apache License (Version 2.0).

See [LICENSE-APACHE](LICENSE-APACHE) and [LICENSE-MIT](LICENSE-MIT) for details.
