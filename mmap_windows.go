// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mmap

import (
	"os"
	"sync"
	"syscall"
)

// mmap on Windows is a two-step process.
// First, we call CreateFileMapping to get a handle.
// Then, we call MapviewToFile to get an actual pointer into memory.
// Because we want to emulate a POSIX-style mmap, we don't want to expose
// the handle -- only the pointer. We also want to return only a byte slice,
// not a struct, so it's convenient to manipulate.

// We keep this map so that we can get back the original handle from the memory address.
var handleLock sync.Mutex
var handleMap = map[uintptr]int32{}

var (
	modkernel32 = loadDll("kernel32.dll")

	procCreateFileMapping = getSysProcAddr(modkernel32, "CreateFileMappingW")
	procMapViewOfFile     = getSysProcAddr(modkernel32, "MapViewOfFile")
	procFlushViewOfFile   = getSysProcAddr(modkernel32, "FlushViewOfFile")
	procUnmapViewOfFile   = getSysProcAddr(modkernel32, "UnmapViewOfFile")
	procVirtualLock       = getSysProcAddr(modkernel32, "VirtualLock")
	procVirtualUnlock     = getSysProcAddr(modkernel32, "VirtualUnlock")
	procVirtualProtect    = getSysProcAddr(modkernel32, "VirtualProtect")
)

func loadDll(fname string) uint32 {
	h, e := syscall.LoadLibrary(fname)
	if e != 0 {
		panic("LoadLibrary failed")
	}
	return h
}

func getSysProcAddr(m uint32, pname string) uintptr {
	p, e := syscall.GetProcAddress(m, pname)
	if e != 0 {
		panic("GetProcAddress failed")
	}
	return uintptr(p)
}

const (
	_PAGE_READONLY          = 0x02
	_PAGE_READWRITE         = 0x04
	_PAGE_WRITECOPY         = 0x08
	_PAGE_EXECUTE_READ      = 0x20
	_PAGE_EXECUTE_READWRITE = 0x40
	_PAGE_EXECUTE_WRITECOPY = 0x80

	_SEC_IMAGE        = 0x1000000
	_SEC_RESERVE      = 0x4000000
	_SEC_COMMIT       = 0x8000000
	_SEC_NOCACHE      = 0x10000000
	_SEC_WRITECOMBINE = 0x40000000
	_SEC_LARGE_PAGES  = 0x80000000

	_FILE_MAP_COPY       = 0x01
	_FILE_MAP_WRITE      = 0x02
	_FILE_MAP_READ       = 0x04
	_FILE_MAP_ALL_ACCESS = 0x000F0000 | _FILE_MAP_COPY | _FILE_MAP_READ | _FILE_MAP_WRITE | 0x08 | 0x10
	_FILE_MAP_EXECUTE    = 0x20
)

func mmap(len int, prot, flags, hfile uintptr, off int64) ([]byte, os.Error) {
	flProtect := uintptr(_PAGE_READONLY)
	dwDesiredAccess := uintptr(_FILE_MAP_READ)
	switch {
	case prot&COPY != 0:
		flProtect = _PAGE_WRITECOPY
		dwDesiredAccess = _FILE_MAP_COPY
	case prot&RDWR != 0:
		flProtect = _PAGE_READWRITE
		dwDesiredAccess = _FILE_MAP_WRITE
	}
	if prot&EXEC != 0 {
		flProtect <<= 4
		dwDesiredAccess |= _FILE_MAP_EXECUTE
	}

	// TODO: Do we need to set some security attributes? It might help portability.
	h, _, errno := syscall.Syscall6(procCreateFileMapping, 6, uintptr(hfile), 0, flProtect, 0, uintptr(len), 0)
	if h == 0 {
		return nil, os.NewSyscallError("CreateFileMapping", int(errno))
	}

	addr, _, errno := syscall.Syscall6(procMapViewOfFile, 5, h, dwDesiredAccess, uintptr(off>>32), uintptr(off&0xFFFFFFFF), uintptr(len), 0)
	if addr == 0 {
		return nil, os.NewSyscallError("MapViewOfFile", int(errno))
	}
	handleLock.Lock()
	handleMap[addr] = int32(h)
	handleLock.Unlock()


	m := MMap{}
	dh := m.header()
	dh.Data = addr
	// TODO: eek, truncation
	dh.Len = int(len)
	dh.Cap = dh.Len

	return m, nil
}

func flush(addr, len uintptr) os.Error {
	_, _, errno := syscall.Syscall(procFlushViewOfFile, 2, addr, len, 0)
	return os.NewSyscallError("FlushViewOfFile", int(errno))
}

func lock(addr, len uintptr) os.Error {
	_, _, errno := syscall.Syscall(procVirtualLock, 2, addr, len, 0)
	return os.NewSyscallError("VirtualLock", int(errno))
}

func unlock(addr, len uintptr) os.Error {
	_, _, errno := syscall.Syscall(procVirtualUnlock, 2, addr, len, 0)
	return os.NewSyscallError("VirtualUnlock", int(errno))
}

func unmap(addr, len uintptr) os.Error {
	flush(addr, len)
	r0, _, errno := syscall.Syscall(procUnmapViewOfFile, 1, addr, 0, 0)
	if r0 == 0 {
		return os.NewSyscallError("UnmapViewOfFile", int(errno))
	}

	handleLock.Lock()
	defer handleLock.Unlock()
	handle, ok := handleMap[addr]
	if !ok {
		// should be impossible; we would've errored above
		return os.NewError("unknown base address")
	}
	handleMap[addr] = 0, false

	e := syscall.CloseHandle(handle)
	return os.NewSyscallError("CloseHandle", e)
}
