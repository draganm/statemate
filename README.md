# StateMate

StateMate is a Go package that provides an efficient and concurrent-safe method for managing state in the form of key-value pairs. The package allows users to append data with an index and read data by index. It uses memory-mapped files for better performance.

## Features

- Memory-mapped files for fast I/O
- Concurrent-safe with built-in locking mechanism
- Allow/disallow index gaps
- Dynamically extend the underlying files as needed

## Requirements

- Go version 1.18 or higher (due to the use of generics)

## Installation

To install StateMate, use `go get`:

```bash
go get -u github.com/draganm/statemate
```

## Usage

### Importing

```go
import "github.com/draganm/statemate"
```

### Initialization

```go
options := statemate.Options{AllowGaps: false}
sm, err := statemate.Open[uint64]("datafile", options)
if err != nil {
    log.Fatal(err)
}
defer sm.Close()
```

### Appending Data

```go
err = sm.Append(1, []byte("data"))
if err != nil {
    log.Fatal(err)
}
```

### Reading Data

```go
err = sm.Read(1, func(data []byte) error {
    fmt.Println(string(data))
    return nil
})
if err != nil {
    log.Fatal(err)
}
```

### Errors

- `ErrIndexMustBeIncreasing`: Raised when the new index is not greater than the last index.
- `ErrIndexGapsAreNotAllowed`: Raised when gaps are detected and `AllowGaps` option is set to `false`.
- `ErrNotFound`: Raised when trying to read an index that doesn't exist.

## License

MIT
