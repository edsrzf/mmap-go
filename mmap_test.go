// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// These tests are adapted from gommap: http://labix.org/gommap
// Copyright (c) 2010, Gustavo Niemeyer <gustavo@niemeyer.net>

package mmap

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

var testData = []byte("0123456789ABCDEF")
var testPath = filepath.Join(os.TempDir(), "testdata")

func init() {
	f := openFile(os.O_RDWR | os.O_CREATE | os.O_TRUNC)
	f.Write(testData)
	f.Close()
}

func openFile(flags int) *os.File {
	f, err := os.OpenFile(testPath, flags, 0644)
	if err != nil {
		panic(err.Error())
	}
	return f
}

func TestUnmap(t *testing.T) {
	f := openFile(os.O_RDONLY)
	defer f.Close()
	mmap, err := Map(f, RDONLY, 0)
	if err != nil {
		t.Errorf("error mapping: %s", err)
	}
	if err := mmap.Unmap(); err != nil {
		t.Errorf("error unmapping: %s", err)
	}
}

func TestReadWrite(t *testing.T) {
	f := openFile(os.O_RDWR)
	defer f.Close()
	mmap, err := Map(f, RDWR, 0)
	if err != nil {
		t.Errorf("error mapping: %s", err)
	}
	defer mmap.Unmap()
	if !bytes.Equal(testData, mmap) {
		t.Errorf("mmap != testData: %q, %q", mmap, testData)
	}

	mmap[9] = 'X'
	mmap.Flush()

	fileData, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("error reading file: %s", err)
	}
	if !bytes.Equal(fileData, []byte("012345678XABCDEF")) {
		t.Errorf("file wasn't modified")
	}

	// leave things how we found them
	mmap[9] = '9'
	mmap.Flush()
}

func TestProtFlagsAndErr(t *testing.T) {
	f := openFile(os.O_RDONLY)
	defer f.Close()
	if _, err := Map(f, RDWR, 0); err == nil {
		t.Errorf("expected error")
	}
}

func TestFlags(t *testing.T) {
	f := openFile(os.O_RDWR)
	defer f.Close()
	mmap, err := Map(f, COPY, 0)
	if err != nil {
		t.Errorf("error mapping: %s", err)
	}
	defer mmap.Unmap()

	mmap[9] = 'X'
	mmap.Flush()

	fileData, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("error reading file: %s", err)
	}
	if !bytes.Equal(fileData, testData) {
		t.Errorf("file was modified")
	}
}

// Test that we can map files from non-0 offsets
// The page size on most Unixes is 4KB, but on Windows it's 64KB
func TestNonZeroOffset(t *testing.T) {
	const pageSize = 65536

	// Create a 2-page sized file
	bigFilePath := filepath.Join(os.TempDir(), "nonzero")
	fileobj, err := os.OpenFile(bigFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err.Error())
	}

	bigData := make([]byte, 2*pageSize, 2*pageSize)
	fileobj.Write(bigData)
	fileobj.Close()

	// Map the first page by itself
	fileobj, err = os.OpenFile(bigFilePath, os.O_RDONLY, 0)
	if err != nil {
		panic(err.Error())
	}
	m, err := MapRegion(fileobj, pageSize, RDONLY, 0, 0)
	if err != nil {
		t.Errorf("error mapping file: %s", err)
	}
	m.Unmap()
	fileobj.Close()

	// Map the second page by itself
	fileobj, err = os.OpenFile(bigFilePath, os.O_RDONLY, 0)
	if err != nil {
		panic(err.Error())
	}
	m, err = MapRegion(fileobj, pageSize, RDONLY, 0, pageSize)
	if err != nil {
		t.Errorf("error mapping file: %s", err)
	}
	err = m.Unmap()
	if err != nil {
		t.Error(err)
	}

	m, err = MapRegion(fileobj, pageSize, RDONLY, 0, 1)
	if err == nil {
		t.Error("expect error because offset is not multiple of page size")
	}

	fileobj.Close()
}

func TestAnonymousMapping(t *testing.T) {
	const size = 4 * 1024

	// Make an anonymous region
	mem, err := MapRegion(nil, size, RDWR, ANON, 0)
	if err != nil {
		t.Fatalf("failed to allocate memory for buffer: %v", err)
	}

	// Check memory writable
	for i := 0; i < size; i++ {
		mem[i] = 0x55
	}

	// And unmap it
	err = mem.Unmap()
	if err != nil {
		t.Fatalf("failed to unmap memory for buffer: %v", err)
	}
}

func TestAnonymousMappingCopy(t *testing.T) {
	const size = 4 * 1024

	// Make an anonymous copy-on-write region
	mem, err := MapRegion(nil, size, COPY, ANON, 0)
	if err != nil {
		t.Fatalf("failed to allocate memory for copy buffer: %v", err)
	}

	for i := 0; i < size; i++ {
		mem[i] = 0xAA
	}

	err = mem.Unmap()
	if err != nil {
		t.Fatalf("failed to unmap memory for copy buffer: %v", err)
	}
}

func TestUserFlagsPreserved(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping linux-specific flag tests on non-linux OS")
	}

	const (
		flagMapNoreserve = 0x4000
		flagMapPopulate  = 0x8000
		flagMapHugetlb   = 0x40000
	)

	// Test mapping a file with MAP_POPULATE flag
	f := openFile(os.O_RDWR)
	defer f.Close()
	m, err := Map(f, RDWR, flagMapPopulate)
	if err != nil {
		t.Fatalf("Map with MAP_POPULATE failed: %v", err)
	}
	if !bytes.Equal(m, testData) {
		t.Fatalf("mmap != testData: %q, %q", m, testData)
	}
	if err := m.Unmap(); err != nil {
		t.Fatalf("failed to unmap MAP_POPULATE region: %v", err)
	}

	// Test mapping a file region with MAP_NORESERVE flag
	mRegion, err := MapRegion(f, len(testData), RDONLY, flagMapNoreserve, 0)
	if err != nil {
		t.Fatalf("MapRegion with MAP_NORESERVE failed: %v", err)
	}
	if !bytes.Equal(mRegion, testData) {
		t.Fatalf("mRegion != testData: %q, %q", mRegion, testData)
	}
	if err := mRegion.Unmap(); err != nil {
		t.Fatalf("failed to unmap MAP_NORESERVE region: %v", err)
	}

	// Test anonymous mapping combined with user flags
	mAnon, err := MapRegion(nil, 4096, RDWR, ANON|flagMapNoreserve, 0)
	if err != nil {
		t.Fatalf("MapRegion with ANON|MAP_NORESERVE failed: %v", err)
	}
	mAnon[0] = 0x5A
	if err := mAnon.Unmap(); err != nil {
		t.Fatalf("failed to unmap ANON|MAP_NORESERVE region: %v", err)
	}

	// Test that flags are not ignored: MAP_HUGETLB requires allocated hugepages in Linux.
	// When passed to unix.Mmap, the kernel fails with EINVAL if hugepages are unavailable.
	// If flags were ignored, this call would succeed.
	_, err = Map(f, RDWR, flagMapHugetlb)
	if err == nil {
		t.Error("expected error when mapping with MAP_HUGETLB without hugepages, but got nil; user flags may have been dropped")
	}
}

