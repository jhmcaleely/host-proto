package main

/*
#include "lfs.h"
#include "block_device.h"
*/
import "C"
import "unsafe"

const LFS_ERR_OK C.int = 0

func fsAddressForBlock(fs_base_address uint32, block, off uint32) uint32 {

	byte_offset := block*PICO_ERASE_PAGE_SIZE + off

	return fs_base_address + byte_offset
}

//export go_bdfs_read
func go_bdfs_read(device *C.struct_block_device, fs_flash_base_address C.uint32_t, block C.lfs_block_t, off C.lfs_off_t, buffer *C.void, size C.lfs_size_t) C.int {

	device_address := fsAddressForBlock(uint32(fs_flash_base_address), uint32(block), uint32(off))

	C.bdRead(device, C.uint32_t(device_address), (*C.uint8_t)(unsafe.Pointer(buffer)), C.size_t(size))

	return LFS_ERR_OK
}

//export go_bdfs_prog_page
func go_bdfs_prog_page(device *C.struct_block_device, fs_flash_base_address C.uint32_t, block C.lfs_block_t, off C.lfs_off_t, buffer *C.void, size C.lfs_size_t) C.int {

	device_address := fsAddressForBlock(uint32(fs_flash_base_address), uint32(block), uint32(off))

	C.bdWrite(device, C.uint32_t(device_address), (*C.uint8_t)(unsafe.Pointer(buffer)), C.size_t(size))

	C.bdDebugPrint(device)

	return LFS_ERR_OK
}

//export go_bdfs_erase_block
func go_bdfs_erase_block(device *C.struct_block_device, fs_flash_base_address C.uint32_t, block C.lfs_block_t) C.int {

	device_address := fsAddressForBlock(uint32(fs_flash_base_address), uint32(block), uint32(0))

	C.bdEraseBlock(device, C.uint32_t(device_address))

	return LFS_ERR_OK
}
