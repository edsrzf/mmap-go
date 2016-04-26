// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux openbsd solaris

package mmap

import (
	"syscall"
)

func msync(addr, len uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_MSYNC, addr, len, syscall.MS_SYNC)
	if errno != 0 {
		return syscall.Errno(errno)
	}
	return nil
}
