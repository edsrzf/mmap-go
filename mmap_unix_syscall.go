// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mmap

import (
	"syscall"
)

func mmap_syscall(len, prot, flags, fd uintptr, off int64) (uintptr, uintptr) {
	ptr, _, errno := syscall.Syscall6(syscall.SYS_MMAP, 0, len, prot,
		flags, fd, uintptr(off))
	return ptr, errno
}
