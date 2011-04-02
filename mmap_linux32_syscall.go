// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mmap

// 32-bit Linux architectures (namely x86 and ARM) have two mmap system calls, both
// of which are incompatible with mmap on other platforms:
//	- syscall.MMAP (old_mmap) expects a single pointer argument
//	- syscall.MMAP2 (mmap2) takes an offset in pages, not bytes
// Thus, 32-bit Linux gets its own special file.
import (
	"syscall"
)

func mmap_syscall(len, prot, flags, fd uintptr, off int64) (uintptr, uintptr) {
	// assuming page size is 4096; the runtime does it, so it should be okay
	if off&0xFFF != 0 {
		return 0, syscall.EINVAL
	}
	off >>= 12
	ptr, _, errno := syscall.Syscall6(syscall.SYS_MMAP2, 0, len, prot,
		flags, fd, uintptr(off))
	return ptr, errno
}
