# StateMate Go Library

## Overview

StateMate is a Go library designed for storing state as key-value pairs with efficient I/O operations and concurrency support.

## Features

- Memory-mapped files for efficient I/O.
- Thread-safe via read-write mutexes.
- Generics support for custom unsigned integer key types.
- Dynamic resizing of data and index files.
- Optional index gap allowance.

## Requirements

- Go 1.18 or higher for generics support.
  
## Installation

Install the package using:

```
go get -u github.com/draganm/statemate
```

## Usage

### Initialize a StateMate instance

```go
options := statemate.Options{ AllowGaps: true }
sm, err := statemate.Open[uint64]("datafile", options)
if err != nil {
    // Handle error
}
```

### Append Data

```go
err := sm.Append(1, []byte("some data"))
if err != nil {
    // Handle error
}
```

### Read Data

```go
err := sm.Read(1, func(data []byte) error {
    // Process data
    return nil
})
if err != nil {
    // Handle error
}
```

### Check if Empty

```go
isEmpty := sm.IsEmpty()
```

### Get Last Index

```go
lastIndex := sm.LastIndex()
```

## Errors

- `ErrIndexMustBeIncreasing`: The provided index must be greater than the last index.
- `ErrIndexGapsAreNotAllowed`: If `AllowGaps` is `false`, indexes must be consecutive.
- `ErrNotFound`: The requested index was not found.

## License

MIT License

