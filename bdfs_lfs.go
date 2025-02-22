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

func newBdFS(device BlockDevice, baseAddr uint32, blockCount uint32) *BdFS {

	var blockfs C.struct_flash_fs
	blockfs.device = device.chandle
	blockfs.fs_flash_base_address = C.uint32_t(baseAddr)

	cfg := BdFS{cfg: newLittleFsConfig(blockCount), FsP: &blockfs, Device: device.chandle}
	return &cfg
}

func (fs *BdFS) init() error {

	C.install_bdfs_hooks(fs.cfg.chandle, fs.FsP)

	return nil
}

func (fs *BdFS) Close() error {
	C.remove_bdfs_hooks(fs.cfg.chandle, fs.FsP)
	return nil
}
