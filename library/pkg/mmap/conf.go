package mmap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	_defReadBytes = 1024 * 1024
)

type Config struct {
	File   *FileConfig
	Map    *MapConfig
	Kernel *KernelConfig
}

type FileConfig struct {
	Path string
	Name string
	Perm int32
	Mode string
}

type MapConfig struct {
	Prot   int
	Flags  int
	Offset int64
	Length int
}

type KernelConfig struct {
	// The size of a read from a buffer
	ReadBytes int
	// The number of routines to process read bytes
	Routines int
}

func (c *Config) Builder() error {
	if c.File == nil {
		return errors.New("required file configure")
	}
	if c.File.Mode == "" {
		c.File.Mode = "r"
	}
	if c.File.Name == "" {
		return errors.New("required filename")
	}
	if c.File.Path != "" {
		di, err := os.Stat(c.File.Path)
		if c.File.Mode == "r" {
			if err != nil {
				return fmt.Errorf("dir error: %s", err.Error())
			}
			if !di.IsDir() {
				return fmt.Errorf("%s not exists", c.File.Path)
			}
			if c.File.Perm == 0 {
				c.File.Perm = 0644
			}
		} else {
			if err == nil && !di.IsDir() {
				return fmt.Errorf("%s already exists and not a directory", c.File.Path)
			}
			if os.IsNotExist(err) {
				if err = os.MkdirAll(c.File.Path, os.FileMode(c.File.Perm)); err != nil {
					return fmt.Errorf("create dir %s error: %s", c.File.Path, err.Error())
				}
			}
			if c.File.Perm == 0 {
				c.File.Perm = 0644
			}
		}
		fpath := filepath.Join(c.File.Path, c.File.Name)
		_, err = os.Stat(fpath)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		}
	}

	if c.Map == nil {
		c.Map = &MapConfig{}
	}
	if c.Map.Offset%int64(os.Getpagesize()) != 0 {
		return fmt.Errorf("offset must be a multiple of the system's page size(%d)", os.Getpagesize())
	}

	if c.Kernel == nil {
		c.Kernel = &KernelConfig{
			ReadBytes: _defReadBytes,
			Routines:  1,
		}
	}
	return nil
}
