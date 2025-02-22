package main

/*
#include "bdfs_lfs.h"
*/
import "C"
import (
	"runtime"
)

// Defines one region of flash to use for a filesystem. The size is a multiple of
// the 4096 byte erase size. We calculate it's location working back from the end of the
// flash device, so that code flashed at the start of the device will not collide.
// Pico's have a 2Mb flash device, so we're looking to be less than 2Mb.

const (
	// 128 blocks will reserve a 512K filsystem - 1/4 of the 2Mb device on a Pico
	FLASHFS_BLOCK_COUNT = 128
	FLASHFS_SIZE_BYTES  = PICO_ERASE_PAGE_SIZE * FLASHFS_BLOCK_COUNT

	// A start location counted back from the end of the device.
	FLASHFS_BASE_ADDR uint32 = PICO_FLASH_BASE_ADDR + PICO_FLASH_SIZE_BYTES - FLASHFS_SIZE_BYTES
)

type BdFS struct {
	cfg              LittleFsConfig
	device           BlockDevice
	flash_fs_chandle *C.struct_flash_fs
	pins             *runtime.Pinner
}

func newBdFS(device BlockDevice, baseAddr uint32, blockCount uint32) *BdFS {

	var blockfs C.struct_flash_fs
	blockfs.device = device.chandle
	blockfs.fs_flash_base_address = C.uint32_t(baseAddr)

	cfg := BdFS{cfg: *newLittleFsConfig(blockCount), flash_fs_chandle: &blockfs, device: device, pins: &runtime.Pinner{}}

	cfg.pins.Pin(cfg.flash_fs_chandle)
	cfg.pins.Pin(cfg.cfg.chandle)
	cfg.pins.Pin(cfg.device.chandle)

	C.install_bdfs_hooks(cfg.cfg.chandle, cfg.flash_fs_chandle)
	cfg.pins.Pin(cfg.cfg.chandle.context)

	return &cfg
}

func (fs BdFS) Close() error {
	fs.pins.Unpin()

	return nil
}
