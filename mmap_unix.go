// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mmap

import (
	"os"
	"syscall"
)

func mmap(len int64, inprot, inflags, fd uintptr, off int64) (uintptr, os.Error) {
	flags := uintptr(_MAP_SHARED)
	prot := uintptr(_PROT_READ)
	switch {
	case inprot&COPY != 0:
		prot |= _PROT_WRITE
		flags = _MAP_PRIVATE
	case inprot&RDWR != 0:
		prot |= _PROT_WRITE
	}
	if inprot&EXEC != 0 {
		prot |= _PROT_EXEC
	}
	if inflags&ANON != 0 {
		flags |= _MAP_ANONYMOUS
	}

	addr, errno := mmap_syscall(uintptr(len), prot, flags, fd, off)
	if errno != 0 {
		return 0, os.Errno(errno)
	}
	return addr, nil
}

func flush(addr, len uintptr) os.Error {
	_, _, errno := syscall.Syscall(syscall.SYS_MSYNC, addr, _MS_SYNC, 0)
	if errno != 0 {
		return os.Errno(errno)
	}
	return nil
}

func lock(addr, len uintptr) os.Error {
	_, _, errno := syscall.Syscall(syscall.SYS_MLOCK, addr, len, 0)
	if errno != 0 {
		return os.Errno(errno)
	}
	return nil
}

func unlock(addr, len uintptr) os.Error {
	_, _, errno := syscall.Syscall(syscall.SYS_MUNLOCK, addr, len, 0)
	if errno != 0 {
		return os.Errno(errno)
	}
	return nil
}

func unmap(addr, len uintptr) os.Error {
	_, _, errno := syscall.Syscall(syscall.SYS_MUNMAP, addr, len, 0)
	if errno != 0 {
		return os.Errno(errno)
	}
	return nil
}
