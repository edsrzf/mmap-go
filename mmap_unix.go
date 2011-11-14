// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mmap

import (
	"os"
	"syscall"
)

func mmap(len int, inprot, inflags, fd uintptr, off int64) ([]byte, os.Error) {
	flags := syscall.MAP_SHARED
	prot := syscall.PROT_READ
	switch {
	case inprot&COPY != 0:
		prot |= syscall.PROT_WRITE
		flags = syscall.MAP_PRIVATE
	case inprot&RDWR != 0:
		prot |= syscall.PROT_WRITE
	}
	if inprot&EXEC != 0 {
		prot |= syscall.PROT_EXEC
	}
	if inflags&ANON != 0 {
		flags |= MAP_ANONYMOUS
	}

	b, errno := syscall.Mmap(int(fd), off, len, prot, flags)
	if errno != 0 {
		return nil, os.Errno(errno)
	}
	return b, nil
}

func flush(addr, len uintptr) os.Error {
	_, _, errno := syscall.Syscall(syscall.SYS_MSYNC, addr, syscall.MS_SYNC, 0)
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
