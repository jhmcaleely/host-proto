package main

/*
#include "bdfs_lfs.h"
*/
import "C"
import (
	"runtime"
)

type BdFS struct {
	cfg *LittleFsConfig
	FsP *C.struct_flash_fs

	Device *C.struct_block_device

	Pins *runtime.Pinner
}

func newBdFS(device BlockDevice, baseAddr uint32, blockCount uint32) *BdFS {

	var blockfs C.struct_flash_fs
	var pins runtime.Pinner
	blockfs.device = device.chandle
	blockfs.fs_flash_base_address = C.uint32_t(baseAddr)

	cfg := BdFS{cfg: newLittleFsConfig(blockCount), FsP: &blockfs, Device: device.chandle, Pins: &pins}

	cfg.Pins.Pin(cfg.FsP)
	cfg.Pins.Pin(cfg.cfg.chandle)
	cfg.Pins.Pin(cfg.Device)

	cfg.init()

	return &cfg
}

func (fs *BdFS) init() error {

	C.install_bdfs_hooks(fs.cfg.chandle, fs.FsP)
	fs.Pins.Pin(fs.cfg.chandle.context)

	return nil
}

func (fs *BdFS) Close() error {
	defer fs.Pins.Unpin()

	C.remove_bdfs_hooks(fs.cfg.chandle, fs.FsP)

	return nil
}
