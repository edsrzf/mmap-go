// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux openbsd solaris netbsd

package mmap

import (
	"golang.org/x/sys/unix"
)

func mmap(len int, inprot, inflags, fd uintptr, off int64) ([]byte, error) {
	flags := unix.MAP_SHARED
	prot := unix.PROT_READ
	switch {
	case inprot&COPY != 0:
		prot |= unix.PROT_WRITE
		flags = unix.MAP_PRIVATE
	case inprot&RDWR != 0:
		prot |= unix.PROT_WRITE
	}
	if inprot&EXEC != 0 {
		prot |= unix.PROT_EXEC
	}
	if inflags&ANON != 0 {
		flags |= unix.MAP_ANON
	}

	b, err := unix.Mmap(int(fd), off, len, prot, flags)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func flush(addr, len uintptr) error {
	_, _, errno := unix.Syscall(unix.SYS_MSYNC, addr, len, unix.MS_SYNC)
	if errno != 0 {
		return unix.Errno(errno)
	}
	return nil
}

func lock(addr, len uintptr) error {
	_, _, errno := unix.Syscall(unix.SYS_MLOCK, addr, len, 0)
	if errno != 0 {
		return unix.Errno(errno)
	}
	return nil
}

func unlock(addr, len uintptr) error {
	_, _, errno := unix.Syscall(unix.SYS_MUNLOCK, addr, len, 0)
	if errno != 0 {
		return unix.Errno(errno)
	}
	return nil
}

func unmap(addr, len uintptr) error {
	_, _, errno := unix.Syscall(unix.SYS_MUNMAP, addr, len, 0)
	if errno != 0 {
		return unix.Errno(errno)
	}
	return nil
}
