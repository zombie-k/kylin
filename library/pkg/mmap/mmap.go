// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file defines the common package interface and contains a little bit of
// factored out logic.

// Package mmap allows mapping files into memory. It tries to provide a simple, reasonably portable interface,
// but doesn't go out of its way to abstract away every little platform detail.
// This specifically means:
//	* forked processes may or may not inherit mappings
//	* a file's timestamp may or may not be updated by writes through mappings
//	* specifying a size larger than the file's actual size can increase the file's size
//	* If the mapped file is being modified by another process while your program's running, don't expect consistent results between platforms
package mmap

import (
	"errors"
	"os"
)

const (
	RDONLY = 0
	RDWR   = 1 << iota
	COPY
	EXEC
)

type MMap []byte

// Map maps an entire file into memory.
func Map(f *os.File, prot, flags int) (MMap, error) {
	return MapRegion(f, -1, prot, flags, 0)
}

// MapRegion maps part of a file into memory.
// The offset parameter must be a multiple of the system's page size.
// If length < 0, the entire file will be mapped.
func MapRegion(f *os.File, length int, prot, flags int, offset int64) (MMap, error) {
	if offset%int64(os.Getpagesize()) != 0 {
		return nil, errors.New("offset parameter must be a multiple of the system's page size")
	}
	fd := f.Fd()
	if length < 0 {
		fi, err := f.Stat()
		if err != nil {
			return nil, err
		}
		length = int(fi.Size())
	}

	return mmap(length, prot, flags, fd, offset)
}

func (m *MMap) Unmap() error {
	err := m.unmap()
	*m = nil
	return err
}

