// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file defines the common package interface and contains a little bit of
// factored out logic.

// Package mmap allows mapping files into memory. It tries to provide a simple, reasonably portable interface,
// but doesn't go out of its way to abstract away every little platform detail.
// This specifically means:
//	* forked processes may or may not inherit mappings
//	* a file's timestamp may or may not be updated by writes through mappings
//	* specifying a size larger than the file's actual size can increase the file's size
//	* If the mapped file is being modified by another process while your program's running, don't expect consistent results between platforms
package mmap

import (
	"os"
	"reflect"
	"unsafe"
)

const (
	// RDONLY maps the memory read-only.
	// Attempts to write to the MMap object will result in undefined behavior.
	RDONLY = 0
	// RDWR maps the memory as read-write. Writes to the MMap object will update the
	// underlying file.
	RDWR = 1 << iota
	// COPY maps the memory as copy-on-write. Writes to the MMap object will affect
	// memory, but the underlying file will remain unchanged.
	COPY
	// If EXEC is set, the mapped memory is marked as executable.
	EXEC
)

const (
	// If the ANON flag is set, the mapped memory will not be backed by a file.
	ANON = 1 << iota
)

// MMap represents a file mapped into memory.
type MMap []byte

// Map maps an entire file into memory.
// If ANON is set in flags, f is ignored.
func Map(f *os.File, prot, flags int) (MMap, os.Error) {
	return MapRegion(f, -1, prot, flags, 0)
}

// MapRegion maps part of a file into memory.
// The offset parameter must be a multiple of the system's page size.
// If length < 0, the entire file will be mapped.
// If ANON is set in flags, f is ignored.
func MapRegion(f *os.File, length int64, prot, flags int, offset int64) (MMap, os.Error) {
	var fd uintptr
	if flags&ANON == 0 {
		fd = uintptr(f.Fd())
		if length < 0 {
			fi, err := f.Stat()
			if err != nil {
				return nil, err
			}
			length = fi.Size
		}
	} else {
		if length <= 0 {
			return nil, os.NewError("anonymous mapping requires non-zero length")
		}
		fd = ^uintptr(0)
	}
	addr, err := mmap(length, uintptr(prot), uintptr(flags), fd, offset)
	if err != nil {
		return nil, err
	}

	m := MMap{}
	dh := m.header()
	dh.Data = addr
	// TODO: eek, truncation
	dh.Len = int(length)
	dh.Cap = dh.Len
	return m, nil
}

func (m *MMap) header() *reflect.SliceHeader {
	return (*reflect.SliceHeader)(unsafe.Pointer(m))
}

// Lock keeps the mapped region in physical memory, ensuring that it will not be
// swapped out.
func (m MMap) Lock() os.Error {
	dh := m.header()
	return lock(dh.Data, uintptr(dh.Len))
}

// Unlock reverses the effect of Lock, allowing the mapped region to potentially
// be swapped out.
// If m is already unlocked, aan error will result.
func (m MMap) Unlock() os.Error {
	dh := m.header()
	return unlock(dh.Data, uintptr(dh.Len))
}

// Flush synchronizes the mapping's contents to the file's contents on disk.
func (m MMap) Flush() os.Error {
	dh := m.header()
	return flush(dh.Data, uintptr(dh.Len))
}

// Unmap deletes the memory mapped region, flushes any remaining changes, and sets
// m to nil.
// Trying to read or write any remaining references to m after Unmap is called will
// result in undefined behavior.
// Unmap should only be called on the slice value that was originally returned from
// a call to Map. Calling Unmap on a derived slice may cause errors.
func (m *MMap) Unmap() os.Error {
	dh := m.header()
	err := unmap(dh.Data, uintptr(dh.Len))
	*m = nil
	return err
}
