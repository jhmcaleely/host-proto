package main

/*
#include "lfs.h"
*/
import "C"
import (
	"errors"
	"runtime"
)

type LittleFs struct {
	lfs C.lfs_t
}

func (fs *LittleFs) mount(cfg *C.struct_lfs_config) error {
	lfsp := &fs.lfs
	var pin runtime.Pinner
	pin.Pin(lfsp)
	defer pin.Unpin()

	result := C.lfs_mount(lfsp, cfg)
	if result < 0 {
		return errors.New("mount failed")
	} else {
		return nil
	}
}

func (fs *LittleFs) unmount() error {
	lfsp := &fs.lfs
	var pin runtime.Pinner
	pin.Pin(lfsp)
	defer pin.Unpin()

	result := C.lfs_unmount(lfsp)
	if result < 0 {
		return errors.New("unmount failed")
	} else {
		return nil
	}
}

func lfsFormat(cfg *C.struct_lfs_config) error {
	var lfs C.lfs_t
	lfsp := &lfs
	var pin runtime.Pinner
	pin.Pin(lfsp)
	defer pin.Unpin()

	result := C.lfs_format(lfsp, cfg)
	if result < 0 {
		return errors.New("format failed")
	} else {
		return nil
	}
}
