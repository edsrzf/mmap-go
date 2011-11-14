# Copyright 2011 Evan Shaw. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.inc

TARG=github.com/edsrzf/mmap-go
GOFILES=\
	mmap.go

GOFILES_freebsd=\
	mmap_darwin.go\
	mmap_unix.go

GOFILES_darwin=\
	mmap_darwin.go\
	mmap_unix.go

GOFILES_linux=\
	mmap_linux.go\
	mmap_unix.go

GOFILES_windows=\
	mmap_windows.go

GOFILES+=$(GOFILES_$(GOOS))
include $(GOROOT)/src/Make.pkg
