package main

/*
#include "lfs.h"
*/
import "C"
import (
	"errors"
	"runtime"
	"unsafe"
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

func (fs *LittleFs) Close() error {
	return fs.unmount()
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

type LfsDir struct {
	Lfs *LittleFs
	Dir C.lfs_dir_t
}

func (dir *LfsDir) Open(name string) error {
	dirp := &dir.Dir
	var pin runtime.Pinner
	pin.Pin(dirp)
	defer pin.Unpin()
	lfsp := &dir.Lfs.lfs
	pin.Pin(lfsp)

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	result := C.lfs_dir_open(lfsp, dirp, cname)
	if result < 0 {
		return errors.New("dir open failed")
	}
	return nil
}

func (dir *LfsDir) Close() error {
	dirp := &dir.Dir
	var pin runtime.Pinner
	pin.Pin(dirp)
	defer pin.Unpin()
	lfsp := &dir.Lfs.lfs
	pin.Pin(lfsp)

	result := C.lfs_dir_close(lfsp, dirp)
	if result < 0 {
		return errors.New("dir close failed")
	}
	return nil
}

func (dir *LfsDir) Read(info *C.struct_lfs_info) (bool, error) {
	dirp := &dir.Dir
	var pin runtime.Pinner
	pin.Pin(dirp)
	defer pin.Unpin()
	lfsp := &dir.Lfs.lfs
	pin.Pin(lfsp)

	result := C.lfs_dir_read(lfsp, dirp, info)
	if result < 0 {
		return false, errors.New("dir read failed")
	} else {
		return result != 0, nil
	}
}
