package statemate

import (
	"fmt"
	"os"

	"github.com/edsrzf/mmap-go"
)

type StateMate struct {
	readOnly mmap.MMap
}

func Open(fileName string) (*StateMate, error) {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat file: %w", err)
	}

	if fi.Size() < 8 {
		err = f.Truncate(8)
		if err != nil {
			return nil, fmt.Errorf("failed extending file to 8 bytes: %w", err)
		}
	}

	readOnly, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("could not create read only mmap: %w", err)
	}

	return &StateMate{readOnly: readOnly}, nil

	// readWrite, err := mmap.Map(f, mmap.RDWR, 0)
}
