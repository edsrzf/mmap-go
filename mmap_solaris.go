// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build solaris

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

func flush(b []byte) error {
	err := unix.Msync(b, unix.MS_SYNC)
	if err != nil {
		return err
	}
	return nil
}

func lock(b []byte) error {
	err := unix.Mlock(b)
	if err != nil {
		return err
	}
	return nil
}

func unlock(b []byte) error {
	err := unix.Munlock(b)
	if err != nil {
		return err
	}
	return nil
}

func unmap(b []byte) error {
	err := unix.Munmap(b)
	if err != nil {
		return err
	}
	return nil
}
