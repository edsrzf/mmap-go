// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include <sys/mman.h>

enum {
	$_PROT_READ = PROT_READ,
	$_PROT_WRITE = PROT_WRITE,
	$_PROT_EXEC = PROT_EXEC,

	$_MAP_ANONYMOUS = MAP_ANONYMOUS,
	$_MAP_SHARED = MAP_SHARED,
	$_MAP_PRIVATE = MAP_PRIVATE,

	$_MS_SYNC = MS_SYNC,
};
