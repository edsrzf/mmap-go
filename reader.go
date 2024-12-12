package mmap

import (
	"bytes"
	"io"
	"runtime"
	"runtime/debug"
)

type FaultReader struct {
	mmap   MMap
	reader *bytes.Reader
}

type addressFault interface {
	runtime.Error
	Addr() uintptr
}

func (f *FaultReader) Len() int {
	return f.reader.Len()
}

func (f *FaultReader) Size() int64 {
	return f.reader.Size()
}

func (f *FaultReader) Read(b []byte) (n int, err error) {
	if fault := f.handleFaults(func() {
		n, err = f.reader.Read(b)
	}); fault != nil {
		err = fault
	}
	return
}

func (f *FaultReader) ReadAt(b []byte, off int64) (n int, err error) {
	if fault := f.handleFaults(func() {
		n, err = f.reader.ReadAt(b, off)
	}); fault != nil {
		err = fault
	}
	return
}

func (f *FaultReader) Seek(offset int64, whence int) (int64, error) {
	return f.reader.Seek(offset, whence)
}

func (f *FaultReader) WriteTo(w io.Writer) (n int64, err error) {
	if fault := f.handleFaults(func() {
		n, err = f.reader.WriteTo(w)
	}); fault != nil {
		err = fault
	}
	return
}

func (f *FaultReader) handleFaults(forFunction func()) (err error) {
	previousSetting := debug.SetPanicOnFault(true)
	defer func() {
		debug.SetPanicOnFault(previousSetting)
		if panicErr := recover(); panicErr != nil {
			fault, isAddressFault := panicErr.(addressFault)
			if !isAddressFault {
				panic(panicErr)
			}
			address := fault.Addr()
			mappedAddr, mappedLen := f.mmap.addrLen()
			if mappedAddr <= address && address < mappedAddr+mappedLen {
				// Fault while reading our data
				err = fault
			} else {
				// Forward panic
				// This is not perfect because we potentially downgraded a runtime fault to a panic, but Golang does not
				// allow triggering a runtime fault directly
				panic(fault)
			}
		}
	}()
	forFunction()
	return nil
}
