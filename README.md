# StateMate

[![test](https://github.com/draganm/statemate/actions/workflows/test.yml/badge.svg)](https://github.com/draganm/statemate/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/draganm/statemate.svg)](https://pkg.go.dev/github.com/draganm/statemate)

StateMate is a small Go library for storing an append-only sequence of byte blobs on disk, each addressed by a strictly increasing unsigned integer index. It suits event logs, state snapshots, replicated log segments and similar data that is written once, in order, and later read back by index.

Entries and their index live in two memory-mapped files, so reads are zero-copy and appends are cheap.

## Features

- **Append-only and index addressed.** Entries are written with an index of any `~uint64` type and read back by that index.
- **Memory-mapped I/O.** Reads hand you a slice straight from the mapped file, without copying.
- **Safe for concurrent use.** A read-write mutex allows many concurrent readers and one writer within a process.
- **Pre-allocation with bounded growth.** Files grow geometrically to avoid resizing on every append, and the data file can be capped with `MaxSize`.
- **Optional gaps.** Indexes must always increase. Consecutive indexes are enforced by default and can be relaxed with `AllowGaps`.
- **Truncate for archiving.** Padding can be dropped so files are exactly as large as the data they hold.
- **Storage statistics.** Logical and physical sizes are exposed for monitoring.
- **Command-line tool.** Inspect, truncate and merge state files without writing code.

## Requirements

Go 1.20 or newer.

## Installation

```
go get github.com/draganm/statemate
```

## Usage

### Open and close

```go
sm, err := statemate.Open[uint64]("events", statemate.Options{})
if err != nil {
    return err
}
defer sm.Close()
```

`Open` creates the data file `events` and the index file `events.idx` if they do not exist, and maps them into memory. The type parameter is the index type. Any type whose underlying type is `uint64` works, so a custom `type EventID uint64` can be used directly.

### Append

```go
err := sm.Append(1, []byte("first"))
```

Indexes must be strictly increasing. The first index of an empty store can be any value. Unless `AllowGaps` is set, every following index must be exactly the previous index plus one.

### Read

```go
err := sm.Read(1, func(data []byte) error {
    fmt.Println(string(data))
    return nil
})
if errors.Is(err, statemate.ErrNotFound) {
    // there is no entry with index 1
}
```

The `data` slice points directly into the memory-mapped file. Do not modify it, and do not keep it after the callback returns. Copy it if you need it later.

The store holds a read lock for the duration of the callback. Keep callbacks short, and never call `Append` or `Truncate` from inside one, because that would deadlock.

### Inspect

```go
sm.IsEmpty()       // true when there are no entries
sm.Count()         // number of entries
sm.GetFirstIndex() // index of the first entry
sm.GetLastIndex()  // index of the last entry
```

`GetFirstIndex` and `GetLastIndex` return `math.MaxUint64` when the store is empty, so check `IsEmpty` first.

### Storage statistics

```go
stats := sm.StorageStats()
fmt.Printf("data: %d of %d bytes used, index: %d of %d bytes used\n",
    stats.DataSize, stats.DataFileSize, stats.IndexSize, stats.IndexFileSize)
```

`DataSize` and `IndexSize` are the bytes actually holding entries. `DataFileSize` and `IndexFileSize` are the sizes of the files on disk, which are usually larger because of pre-allocation.

### Truncate

```go
err := sm.Truncate()
```

Drops the pre-allocated padding so both files are exactly as large as their contents. Call it before archiving or copying a store. The next `Append` grows the files again as needed.

### Complete example

```go
package main

import (
    "fmt"
    "log"

    "github.com/draganm/statemate"
)

func main() {
    sm, err := statemate.Open[uint64]("events", statemate.Options{})
    if err != nil {
        log.Fatal(err)
    }
    defer sm.Close()

    for i, msg := range []string{"created", "updated", "deleted"} {
        if err := sm.Append(uint64(i), []byte(msg)); err != nil {
            log.Fatal(err)
        }
    }

    for i := sm.GetFirstIndex(); i <= sm.GetLastIndex(); i++ {
        err := sm.Read(i, func(data []byte) error {
            fmt.Printf("%d: %s\n", i, data)
            return nil
        })
        if err != nil {
            log.Fatal(err)
        }
    }

    stats := sm.StorageStats()
    fmt.Printf("%d entries, %d bytes of data\n", sm.Count(), stats.DataSize)
}
```

## Options

| Option      | Default         | Meaning                                                                                                                                                                                                                  |
| ----------- | --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `AllowGaps` | `false`         | When `false`, each index must be exactly the previous index plus one. When `true`, any larger index is accepted.                                                                                                        |
| `MaxSize`   | `0` (unlimited) | Upper bound in bytes for the data file. `Append` returns `ErrNotEnoughSpace` when an entry would not fit within the bound, and `Open` fails if the existing data file is already larger. The index file is not limited. |

## Errors

| Error                       | Returned by | When                                                                  |
| --------------------------- | ----------- | --------------------------------------------------------------------- |
| `ErrIndexMustBeIncreasing`  | `Append`    | The index is less than or equal to the last index in the store.       |
| `ErrIndexGapsAreNotAllowed` | `Append`    | `AllowGaps` is `false` and the index is not the last index plus one.  |
| `ErrNotEnoughSpace`         | `Append`    | Growing the data file to fit the entry would exceed `MaxSize`.        |
| `ErrNotFound`               | `Read`      | There is no entry with the requested index.                           |

All errors can be tested with `errors.Is`.

## Concurrency

A `StateMate` is safe for concurrent use by multiple goroutines. Readers run in parallel with each other and are blocked only while an `Append` or `Truncate` is in progress.

Coordination happens in memory, so a store must be opened by a single process at a time. Opening the same files from two processes, or twice within one process, and writing through both will corrupt the store.

## On-disk format

A store consists of two files.

- **Data file** (`<name>`): the entries concatenated back to back, without any framing, followed by pre-allocated padding. The file is never smaller than one byte, because mapping an empty file fails on macOS.
- **Index file** (`<name>.idx`): an 8-byte big-endian entry count, followed by one 16-byte record per entry. Each record holds the 8-byte big-endian index and the 8-byte big-endian end offset of the entry's bytes in the data file. An entry starts where the previous one ends, and the first entry starts at offset 0.

When an append does not fit, the file is grown to 1.5 times the required size while it is under 1 GiB, and to the next whole GiB above that. `MaxSize` caps the growth of the data file.

## Command-line tool

```
go install github.com/draganm/statemate/cmd/statemate@latest
```

All commands accept their arguments as flags or as environment variables.

| Command                                                                          | Environment                   | Purpose                                                                                                                                                                           |
| -------------------------------------------------------------------------------- | ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `statemate info --state <file>`                                                  | `STATE`                       | Prints the first index, last index and entry count of a store.                                                                                                                    |
| `statemate truncate --state <file>`                                              | `STATE`                       | Drops the pre-allocated padding from the data and index files.                                                                                                                    |
| `statemate merge --input-files <a> --input-files <b> ... --output-file <merged>` | `INPUT_FILES`, `OUTPUT_FILE`  | Copies several gap-free stores into a new one. The inputs are sorted by first index, and each must end exactly one index before the next begins, otherwise the command fails.    |

## Development

```
go test -race ./...
```

Continuous integration runs `gofmt`, `go vet`, `go build` and the test suite with the race detector on Linux and macOS, against Go 1.20 and the latest stable release.

## License

MIT. See [LICENCE](LICENCE).
