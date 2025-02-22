package main

/*
#include "block_device.h"
*/
import "C"

import (
	"fmt"
)

const PICO_FLASH_BASE_ADDR uint32 = 0x10000000
const PICO_FLASH_SIZE_BYTES = (2 * 1024 * 1024)
const PICO_ERASE_PAGE_SIZE = 4096
const PICO_PROG_PAGE_SIZE = 256

const PICO_DEVICE_BLOCK_COUNT = PICO_FLASH_SIZE_BYTES / PICO_ERASE_PAGE_SIZE
const PICO_FLASH_PAGE_PER_BLOCK = PICO_ERASE_PAGE_SIZE / PICO_PROG_PAGE_SIZE

type BlockDevice struct {
	chandle *C.struct_block_device
}

func newBlockDevice() BlockDevice {
	cdevice := C.bdCreate(C.uint32_t(PICO_FLASH_BASE_ADDR))
	device := BlockDevice{chandle: cdevice}
	return device
}

func (bd BlockDevice) Close() error {
	C.bdDestroy(bd.chandle)
	return nil
}

func (bd BlockDevice) CountPages() uint32 {
	count := uint32(0)

	for b := C.uint32_t(0); b < PICO_DEVICE_BLOCK_COUNT; b++ {
		for p := C.uint32_t(0); p < PICO_FLASH_PAGE_PER_BLOCK; p++ {
			if C.bdPagePresent(bd.chandle, b, p) {
				count++
			}
		}
	}

	return count
}

func bdDebugPrint(bd *C.struct_block_device) {
	for b := C.uint32_t(0); b < PICO_DEVICE_BLOCK_COUNT; b++ {
		for p := C.uint32_t(0); p < PICO_FLASH_PAGE_PER_BLOCK; p++ {
			if C.bdPagePresent(bd, b, p) {
				fmt.Printf("Page [%v, %v]: 0x%08x\n", b, p, C.bdTargetAddress(bd, b, p))
			}
		}
	}
}
