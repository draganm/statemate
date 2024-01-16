package statemate

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"sync"

	"github.com/edsrzf/mmap-go"
)

type StateMate[T ~uint64] struct {
	options Options

	readOnlyData  mmap.MMap
	data          *os.File
	readOnlyIndex mmap.MMap
	index         *os.File

	mu *sync.RWMutex
}

type Options struct {
	AllowGaps bool
	MaxSize   uint64
}

func (o Options) GetMaxSize() uint64 {
	if o.MaxSize == 0 {
		return math.MaxUint64
	}

	return o.MaxSize
}

func Open[T ~uint64](dataFileName string, options Options) (*StateMate[T], error) {
	dataFile, err := os.OpenFile(dataFileName, os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	{
		fi, err := dataFile.Stat()
		if err != nil {
			return nil, fmt.Errorf("could not stat file: %w", err)
		}

		if fi.Size() < 1 {
			err = dataFile.Truncate(1)
			if err != nil {
				return nil, fmt.Errorf("failed extending index file to 8 bytes: %w", err)
			}
		}

		if uint64(fi.Size()) > options.GetMaxSize() {
			return nil, fmt.Errorf("file size %d is larger than max size %d", fi.Size(), options.MaxSize)
		}
	}

	indexFileName := dataFileName + ".idx"

	indexFile, err := os.OpenFile(indexFileName, os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	{

		fi, err := indexFile.Stat()
		if err != nil {
			return nil, fmt.Errorf("could not stat file: %w", err)
		}

		if fi.Size() < 16 {
			err = indexFile.Truncate(8)
			if err != nil {
				return nil, fmt.Errorf("failed extending index file to 8 bytes: %w", err)
			}
		}
	}

	readOnlyData, err := mmap.Map(dataFile, mmap.RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("could not create read only data mmap: %w", err)
	}

	readOnlyIndex, err := mmap.Map(indexFile, mmap.RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("could not create read only index mmap: %w", err)
	}

	return &StateMate[T]{
		options:       options,
		readOnlyData:  readOnlyData,
		data:          dataFile,
		readOnlyIndex: readOnlyIndex,
		index:         indexFile,
		mu:            &sync.RWMutex{},
	}, nil

}

func (sm *StateMate[T]) Close() error {

	return errors.Join(
		sm.readOnlyData.Unmap(),
		sm.data.Close(),
		sm.readOnlyIndex.Unmap(),
		sm.index.Close(),
	)

}

const gByte = 1024 * 1024 * 1024

var ErrNotEnoughSpace = errors.New("not enough space")

func calculateNewSize(currentSize uint64, spaceAvailable uint64, spaceNeeded uint64, maxSize uint64) (uint64, error) {

	newSize := uint64(0)
	if currentSize+(spaceNeeded-spaceAvailable) < gByte {
		newSize = (currentSize + spaceNeeded) * 15 / 10
	} else {
		newSizeInGB := (currentSize + spaceNeeded) / gByte
		newSize = (newSizeInGB + 1) * gByte
	}

	if newSize > maxSize {
		newSize = maxSize
	}

	if currentSize+(spaceNeeded-spaceAvailable) > newSize {
		return 0, ErrNotEnoughSpace
	}

	return newSize, nil
}

var ErrIndexMustBeIncreasing = errors.New("index must be increasing")
var ErrIndexGapsAreNotAllowed = errors.New("index gaps are not allowed")

func (sm *StateMate[T]) Append(index T, data []byte) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	count := binary.BigEndian.Uint64(sm.readOnlyIndex[:8])

	endOfLastData := uint64(0)
	if count > 0 {
		endOfLastData = binary.BigEndian.Uint64(sm.readOnlyIndex[8:][(count-1)*16+8:])
		lastIndex := binary.BigEndian.Uint64(sm.readOnlyIndex[8:][(count-1)*16:])
		if T(lastIndex) >= index {
			return ErrIndexMustBeIncreasing
		}

		if T(lastIndex)+1 != index && !sm.options.AllowGaps {
			return ErrIndexGapsAreNotAllowed
		}

	}

	available := len(sm.readOnlyData) - int(endOfLastData)

	if available <= len(data) {
		newSize, err := calculateNewSize(endOfLastData, uint64(available), uint64(len(data)), sm.options.GetMaxSize())
		if err != nil {
			return err
		}

		err = sm.data.Truncate(int64(newSize))
		if err != nil {
			return fmt.Errorf("could not truncate data file to new size %d: %w", newSize, err)
		}
		err = sm.readOnlyData.Unmap()
		if err != nil {
			return fmt.Errorf("could not unmap data mmap: %w", err)
		}
		readOnlyData, err := mmap.Map(sm.data, mmap.RDONLY, 0)
		if err != nil {
			return fmt.Errorf("could not create resized read only data mmap: %w", err)
		}

		sm.readOnlyData = readOnlyData

	}

	sizeOfIndex := (count * 16) + 8
	availableForIndex := len(sm.readOnlyIndex) - int(sizeOfIndex)

	if availableForIndex < 16 {
		newSize, err := calculateNewSize(sizeOfIndex, uint64(availableForIndex), 16, math.MaxUint64)
		if err != nil {
			return err
		}
		err = sm.index.Truncate(int64(newSize))
		if err != nil {
			return fmt.Errorf("could not truncate index file to new size %d: %w", newSize, err)
		}
		err = sm.readOnlyIndex.Unmap()
		if err != nil {
			return fmt.Errorf("could not unmap index mmap: %w", err)
		}
		readOnlyIndex, err := mmap.Map(sm.index, mmap.RDONLY, 0)
		if err != nil {
			return fmt.Errorf("could not create resized read only data mmap: %w", err)
		}

		sm.readOnlyIndex = readOnlyIndex

	}

	dataWriteMap, err := mmap.Map(sm.data, mmap.RDWR, 0)
	if err != nil {
		return fmt.Errorf("could not create data RW mmap: %w", err)
	}

	copy(dataWriteMap[endOfLastData:], data)

	err = dataWriteMap.Unmap()
	if err != nil {
		return fmt.Errorf("could not unmap data RW map: %w", err)
	}

	endOfLastData += uint64(len(data))

	indexWriteMap, err := mmap.Map(sm.index, mmap.RDWR, 0)
	if err != nil {
		return fmt.Errorf("could not create index RW mmap: %w", err)
	}

	binary.BigEndian.PutUint64(indexWriteMap[sizeOfIndex:], uint64(index))
	binary.BigEndian.PutUint64(indexWriteMap[sizeOfIndex+8:], endOfLastData)

	binary.BigEndian.PutUint64(indexWriteMap, count+1)

	err = indexWriteMap.Unmap()
	if err != nil {
		return fmt.Errorf("could not unmap index RW map: %w", err)
	}

	return nil
}

// Truncate removes all the padding at the end of the data and index files.
// This method should be called before the state should be archived.
func (sm *StateMate[T]) Truncate() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	count := binary.BigEndian.Uint64(sm.readOnlyIndex[:8])

	endOfLastData := uint64(0)
	if count > 0 {
		endOfLastData = binary.BigEndian.Uint64(sm.readOnlyIndex[8:][(count-1)*16+8:])
	}

	available := len(sm.readOnlyData) - int(endOfLastData)

	if available > 0 {
		if endOfLastData == 0 {
			// reduce file to 1 byte instead of 0
			// mmapping empty files in macOS leads to mmap: invalid argument
			endOfLastData = 1
		}
		err := sm.data.Truncate(int64(endOfLastData))
		if err != nil {
			return fmt.Errorf("could not truncate data file to new size %d: %w", endOfLastData, err)
		}
		err = sm.readOnlyData.Unmap()
		if err != nil {
			return fmt.Errorf("could not unmap data mmap: %w", err)
		}
		readOnlyData, err := mmap.Map(sm.data, mmap.RDONLY, 0)
		if err != nil {
			return fmt.Errorf("could not create resized read only data mmap: %w", err)
		}

		sm.readOnlyData = readOnlyData
	}

	sizeOfIndex := (count * 16) + 8
	availableForIndex := len(sm.readOnlyIndex) - int(sizeOfIndex)

	if availableForIndex > 0 {
		err := sm.index.Truncate(int64(sizeOfIndex))
		if err != nil {
			return fmt.Errorf("could not truncate index file to new size %d: %w", sizeOfIndex, err)
		}
		err = sm.readOnlyIndex.Unmap()
		if err != nil {
			return fmt.Errorf("could not unmap index mmap: %w", err)
		}
		readOnlyIndex, err := mmap.Map(sm.index, mmap.RDONLY, 0)
		if err != nil {
			return fmt.Errorf("could not create resized read only data mmap: %w", err)
		}

		sm.readOnlyIndex = readOnlyIndex

	}

	return nil
}

var ErrNotFound = errors.New("not found")

func (sm *StateMate[T]) Read(index T, fn func(data []byte) error) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	count := binary.BigEndian.Uint64(sm.readOnlyIndex[:8])

	searchSlice := sm.readOnlyIndex[8:]

	indexOf := func(n int) T {
		return T(binary.BigEndian.Uint64(searchSlice[n*16:]))
	}

	indexPos, found := sort.Find(int(count), func(i int) int {
		iind := indexOf(i)
		if index < iind {
			return -1
		}
		if index > iind {
			return 1
		}

		return 0

	})

	if !found {
		return ErrNotFound
	}

	endPos := binary.BigEndian.Uint64(searchSlice[indexPos*16+8:])
	startPos := uint64(0)
	if indexPos != 0 {
		startPos = binary.BigEndian.Uint64(searchSlice[(indexPos-1)*16+8:])
	}

	return fn(sm.readOnlyData[startPos:endPos])

}

func (sm *StateMate[T]) IsEmpty() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := binary.BigEndian.Uint64(sm.readOnlyIndex[:8])

	return count == 0

}

func (sm *StateMate[T]) GetLastIndex() T {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := binary.BigEndian.Uint64(sm.readOnlyIndex[:8])

	if count == 0 {
		return T(uint64(math.MaxUint64))
	}

	return T(binary.BigEndian.Uint64(sm.readOnlyIndex[8:][(count-1)*16:]))

}

func (sm *StateMate[T]) GetFirstIndex() T {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := binary.BigEndian.Uint64(sm.readOnlyIndex[:8])

	if count == 0 {
		return T(uint64(math.MaxUint64))
	}

	return T(binary.BigEndian.Uint64(sm.readOnlyIndex[8:]))

}

func (sm *StateMate[T]) Count() uint64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := binary.BigEndian.Uint64(sm.readOnlyIndex[:8])

	return count

}
