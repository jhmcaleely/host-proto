package main

/*
#include "bdfs_lfs.h"
*/
import "C"

type BdFS struct {
	cfg *LittleFsConfig
	FsP *C.struct_flash_fs

	Device *C.struct_block_device
}

func newBdFS(device *C.struct_block_device, blockCount uint32) *BdFS {
	cfg := BdFS{cfg: newLittleFsConfig(blockCount)}

	var blockfs C.struct_flash_fs
	cfg.FsP = &blockfs

	cfg.Device = device

	return &cfg
}

func (fs *BdFS) init(baseAddr uint32) error {

	C.install_bdfs_hooks(fs.cfg.chandle, fs.FsP, fs.Device, C.uint32_t(baseAddr))

	return nil
}

func (fs *BdFS) Close() error {
	C.remove_bdfs_hooks(fs.cfg.chandle, fs.FsP)
	return nil
}
