package main

/*
#include "bdfs_lfs.h"
*/
import "C"
import (
	"runtime"
)

type BdFS struct {
	cfg              LittleFsConfig
	flash_fs_chandle *C.struct_flash_fs
	bd_chandle       *C.struct_block_device
	pins             *runtime.Pinner
}

func newBdFS(device BlockDevice, baseAddr uint32, blockCount uint32) *BdFS {

	var blockfs C.struct_flash_fs
	blockfs.device = device.chandle
	blockfs.fs_flash_base_address = C.uint32_t(baseAddr)

	cfg := BdFS{cfg: *newLittleFsConfig(blockCount), flash_fs_chandle: &blockfs, bd_chandle: device.chandle, pins: &runtime.Pinner{}}

	cfg.pins.Pin(cfg.flash_fs_chandle)
	cfg.pins.Pin(cfg.cfg.chandle)
	cfg.pins.Pin(cfg.bd_chandle)

	C.install_bdfs_hooks(cfg.cfg.chandle, cfg.flash_fs_chandle)
	cfg.pins.Pin(cfg.cfg.chandle.context)

	return &cfg
}

func (fs BdFS) Close() error {
	fs.pins.Unpin()

	return nil
}
