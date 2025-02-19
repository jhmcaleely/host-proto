package main

/*
#include "lfs.h"
#include "bdfs_lfs.h"
#include "block_device.h"
*/
import "C"
import "unsafe"

const LFS_ERR_OK C.int = 0

func fsAddressForBlock(fs_base_address C.uint32_t, block C.lfs_block_t, off C.lfs_off_t) C.uint32_t {

	byte_offset := block*PICO_ERASE_PAGE_SIZE + off

	return fs_base_address + byte_offset
}

//export go_bdfs_read
func go_bdfs_read(fs *C.struct_flash_fs, block C.lfs_block_t, off C.lfs_off_t, buffer *C.void, size C.lfs_size_t) C.int {

	device_address := fsAddressForBlock(fs.fs_flash_base_address, block, off)

	C.bdRead(fs.device, device_address, (*C.uint8_t)(unsafe.Pointer(buffer)), C.size_t(size))

	return LFS_ERR_OK
}

//export go_bdfs_prog_page
func go_bdfs_prog_page(fs *C.struct_flash_fs, block C.lfs_block_t, off C.lfs_off_t, buffer *C.void, size C.lfs_size_t) C.int {

	device_address := fsAddressForBlock(fs.fs_flash_base_address, block, off)

	C.bdWrite(fs.device, device_address, (*C.uint8_t)(unsafe.Pointer(buffer)), C.size_t(size))

	C.bdDebugPrint(fs.device)

	return LFS_ERR_OK
}

//export go_bdfs_erase_block
func go_bdfs_erase_block(fs *C.struct_flash_fs, block C.lfs_block_t) C.int {

	device_address := fsAddressForBlock(fs.fs_flash_base_address, block, 0)

	C.bdEraseBlock(fs.device, device_address)

	return LFS_ERR_OK
}
