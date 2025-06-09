package mmap

import (
	"errors"
	"io"
	"os"
	"sync"
	"unsafe"
)

// WASM doesn't support traditional memory mapping, so we simulate it
// by reading the entire file into memory.

type wasmMapping struct {
	data        []byte
	file        *os.File
	writable    bool
	copyOnWrite bool
	offset      int64
}

var (
	mappings   = make(map[uintptr]*wasmMapping)
	mappingsMu sync.Mutex
)

func mmap(length int, prot, flags, fd uintptr, offset int64) ([]byte, error) {
	// WASM doesn't support anonymous mappings in the traditional sense
	if flags&ANON != 0 {
		// For anonymous mappings, just allocate memory
		data := make([]byte, length)
		mappingsMu.Lock()
		mappings[uintptr(unsafe.Pointer(&data[0]))] = &wasmMapping{
			data:     data,
			writable: prot&RDWR != 0 || prot&COPY != 0,
		}
		mappingsMu.Unlock()
		return data, nil
	}

	// For file-backed mappings, we need to read the file
	file := os.NewFile(fd, "")
	if file == nil {
		return nil, errors.New("invalid file descriptor")
	}

	// Check if file is read-only when RDWR is requested
	if prot&RDWR != 0 {
		// Try a test write to check permissions
		testPos, _ := file.Seek(0, io.SeekCurrent)
		_, err := file.Write([]byte{})
		if err != nil {
			return nil, err
		}
		file.Seek(testPos, io.SeekStart)
	}

	// Seek to the offset
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	// Get file size if length is -1
	if length < 0 {
		stat, err := file.Stat()
		if err != nil {
			return nil, err
		}
		length = int(stat.Size()) - int(offset)
	}

	// Read the data
	data := make([]byte, length)
	n, err := io.ReadFull(file, data)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	// If we read less than requested, that's OK for files
	if n < length {
		data = data[:n]
	}

	// Reset file position to beginning for subsequent reads
	file.Seek(0, io.SeekStart)

	// Store mapping info
	mappingsMu.Lock()
	mappings[uintptr(unsafe.Pointer(&data[0]))] = &wasmMapping{
		data:        data,
		file:        file,
		writable:    prot&RDWR != 0 && prot&COPY == 0,
		copyOnWrite: prot&COPY != 0,
		offset:      offset,
	}
	mappingsMu.Unlock()

	return data, nil
}

func (m MMap) flush() error {
	if len(m) == 0 {
		return nil
	}

	mappingsMu.Lock()
	mapping, ok := mappings[uintptr(unsafe.Pointer(&m[0]))]
	mappingsMu.Unlock()

	if !ok {
		return errors.New("mapping not found")
	}

	// Only flush if writable and backed by a file and not copy-on-write
	if mapping.writable && mapping.file != nil && !mapping.copyOnWrite {
		// Save current position
		savedPos, _ := mapping.file.Seek(0, io.SeekCurrent)

		if _, err := mapping.file.Seek(mapping.offset, io.SeekStart); err != nil {
			return err
		}
		if _, err := mapping.file.Write(m); err != nil {
			return err
		}
		if err := mapping.file.Sync(); err != nil {
			return err
		}

		// Restore position
		mapping.file.Seek(savedPos, io.SeekStart)
		return nil
	}

	return nil
}

func (m MMap) lock() error {
	// WASM doesn't support memory locking
	return nil
}

func (m MMap) unlock() error {
	// WASM doesn't support memory locking
	return nil
}

func (m MMap) unmap() error {
	if len(m) == 0 {
		return nil
	}

	// Flush changes first
	if err := m.flush(); err != nil {
		return err
	}

	mappingsMu.Lock()
	defer mappingsMu.Unlock()

	ptr := uintptr(unsafe.Pointer(&m[0]))
	_, ok := mappings[ptr]
	if !ok {
		return errors.New("mapping not found")
	}

	// Don't close the file here - the caller is responsible for closing it
	// This matches the behavior of the native mmap implementation where
	// unmapping doesn't close the underlying file descriptor

	// Remove from mappings
	delete(mappings, ptr)

	return nil
}
