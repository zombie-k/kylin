package mmap

import (
	"golang.org/x/sys/unix"
)

func mmap(length int, inprot, inflags int, fd uintptr, offset int64) ([]byte, error) {
	flags := unix.MAP_SHARED
	prot := unix.PROT_READ
	switch {
	case inprot&COPY != 0:
		prot = prot | unix.PROT_WRITE
		flags = unix.MAP_PRIVATE
	case inprot&RDWR != 0:
		prot = prot | unix.PROT_WRITE
	}
	if inprot&EXEC != 0 {
		prot = prot | unix.PROT_EXEC
	}
	b, err := unix.Mmap(int(fd), offset, length, prot, flags)
	if err != nil {
		return nil, err
	}
	return b, err
}
