package main

/*
#include "block_device.h"
*/
import "C"

import (
	"fmt"
)

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

	for b := uint32(0); b < PICO_DEVICE_BLOCK_COUNT; b++ {
		for p := uint32(0); p < PICO_FLASH_PAGE_PER_BLOCK; p++ {
			if bd.PagePresent(b, p) {
				count++
			}
		}
	}

	return count
}

func (bd BlockDevice) DebugPrint() {
	for b := uint32(0); b < PICO_DEVICE_BLOCK_COUNT; b++ {
		for p := uint32(0); p < PICO_FLASH_PAGE_PER_BLOCK; p++ {
			if bd.PagePresent(b, p) {
				fmt.Printf("Page [%v, %v]: 0x%08x\n", b, p, bd.TargetAddress(b, p))
			}
		}
	}
}

func (bd BlockDevice) PagePresent(block, page uint32) bool {
	return C.bdPagePresent(bd.chandle, C.uint32_t(block), C.uint32_t(page)) != 0
}

func storageOffset(block, page uint32) uint32 {
	return block*PICO_ERASE_PAGE_SIZE + page*PICO_PROG_PAGE_SIZE
}

func (bd BlockDevice) TargetAddress(block, page uint32) uint32 {
	return uint32(C.bdBaseAddress(bd.chandle)) + storageOffset(block, page)
}

func (bd BlockDevice) IsBlockStart(targetAddr uint32) bool {
	return (((targetAddr - uint32(C.bdBaseAddress(bd.chandle))) % PICO_ERASE_PAGE_SIZE) == 0)
}

func (bd BlockDevice) getDeviceBlockNo(address uint32) uint32 {
	return (address - uint32(C.bdBaseAddress(bd.chandle))) / PICO_ERASE_PAGE_SIZE
}

func (bd BlockDevice) EraseBlock(address uint32) {

	block := bd.getDeviceBlockNo(address)

	C._bdEraseBlock(bd.chandle, C.uint32_t(block))
}
