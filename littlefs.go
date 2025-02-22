package main

/*
#include "stdlib.h"
#include "lfs.h"
*/
import "C"
import (
	"errors"
	"io"
	"runtime"
	"unsafe"
)

type LittleFsConfig struct {
	chandle *C.struct_lfs_config
}

func newLittleFsConfig(blockCount uint32) *LittleFsConfig {
	var ccfg C.struct_lfs_config

	// block device configuration
	ccfg.read_size = 1
	ccfg.prog_size = PICO_PROG_PAGE_SIZE
	ccfg.block_size = PICO_ERASE_PAGE_SIZE

	// the number of blocks we use for a flash fs.
	// Can be zero if we can read it from the fs.
	ccfg.block_count = C.lfs_size_t(blockCount)

	// cache needs to be a multiple of the programming page size.
	ccfg.cache_size = ccfg.prog_size * 1

	ccfg.lookahead_size = 16
	ccfg.block_cycles = 500

	cfg := LittleFsConfig{chandle: &ccfg}
	return &cfg
}

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

type LfsFile struct {
	Lfs  *LittleFs
	File *C.lfs_file_t
}

func newLfsFile(lfs *LittleFs) *LfsFile {
	var f C.lfs_file_t
	var lf = LfsFile{Lfs: lfs, File: &f}

	return &lf
}

func (file LfsFile) Open(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var pin runtime.Pinner
	pin.Pin(file.File)
	defer pin.Unpin()
	lfsp := &file.Lfs.lfs
	pin.Pin(lfsp)

	oflags := C.int(C.LFS_O_RDWR | C.LFS_O_CREAT)

	result := C.lfs_file_open(lfsp, file.File, cname, oflags)
	if result < 0 {
		return errors.New("file open error")
	}
	return nil
}

func (file LfsFile) Close() error {
	var pin runtime.Pinner
	pin.Pin(file.File)
	defer pin.Unpin()
	lfsp := &file.Lfs.lfs
	pin.Pin(lfsp)

	result := C.lfs_file_close(lfsp, file.File)
	if result < 0 {
		return errors.New("file close error")
	}
	return nil
}

func (file LfsFile) Write(data []byte) (int, error) {

	var pin runtime.Pinner
	pin.Pin(file.File)
	defer pin.Unpin()
	lfsp := &file.Lfs.lfs
	pin.Pin(lfsp)

	pin.Pin(file.Lfs)

	cdata := C.CBytes(data)
	defer C.free(cdata)

	result := C.lfs_file_write(lfsp, file.File, cdata, C.lfs_size_t(len(data)))
	if result < 0 {
		return 0, errors.New("write failed")
	}
	return int(result), nil
}

func (file LfsFile) Read(data []byte) (int, error) {

	var pin runtime.Pinner
	pin.Pin(file.File)
	defer pin.Unpin()

	cdata := C.CBytes(data)
	defer C.free(cdata)

	result := C.lfs_file_read(&file.Lfs.lfs, file.File, cdata, C.lfs_size_t(len(data)))
	if result < 0 {
		return 0, errors.New("read failed")
	} else if result == 0 {
		return 0, io.EOF
	} else {
		copy(data, C.GoBytes(cdata, result))
		return int(result), nil
	}
}

func (file LfsFile) Rewind() error {

	var pin runtime.Pinner
	pin.Pin(file.File)
	defer pin.Unpin()
	lfsp := &file.Lfs.lfs
	pin.Pin(lfsp)

	result := C.lfs_file_rewind(lfsp, file.File)
	if result < 0 {
		return errors.New("rewind failed")
	}
	return nil
}
