# Copyright 2011 Evan Shaw. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.inc

TARG=github.com/edsrzf/mmap-go
GOFILES=\
	const_unix.go\
	mmap.go\

GOFILES_freebsd=\
	mmap_unix.go\
	mmap_unix_syscall.go

GOFILES_darwin=\
	mmap_unix.go\
	mmap_unix_syscall.go

GOFILES_linux=\
	mmap_unix.go
ifeq ($(GOARCH),amd64)
GOFILES_linux+=\
	mmap_unix_syscall.go
else
GOFILES_linux+=\
	mmap_linux32_syscall.go
endif

GOFILES_windows=\
	mmap_windows.go

GOFILES+=$(GOFILES_$(GOOS))
include $(GOROOT)/src/Make.pkg

const_unix.go: const_unix.c
	godefs -gmmap const_unix.c > const_unix.go
