package main

/*
#include "lfs.h"
#include "block_device.h"
*/
import "C"
import "unsafe"

//export go_bdfs_read
func go_bdfs_read(flash_fs C.uintptr_t, block C.lfs_block_t, off C.lfs_off_t, buffer *C.void, size C.lfs_size_t) C.int {

	fs := (*FlashFS)(unsafe.Pointer(uintptr(flash_fs)))
	device_address := fs.AddressForBlock(uint32(block), uint32(off))

	C.bdRead(fs.device.chandle, C.uint32_t(device_address), (*C.uint8_t)(unsafe.Pointer(buffer)), C.size_t(size))

	return C.LFS_ERR_OK
}

//export go_bdfs_prog_page
func go_bdfs_prog_page(flash_fs C.uintptr_t, block C.lfs_block_t, off C.lfs_off_t, buffer *C.void, size C.lfs_size_t) C.int {

	fs := (*FlashFS)(unsafe.Pointer(uintptr(flash_fs)))
	device_address := fs.AddressForBlock(uint32(block), uint32(off))

	C.bdWrite(fs.device.chandle, C.uint32_t(device_address), (*C.uint8_t)(unsafe.Pointer(buffer)), C.size_t(size))

	fs.device.DebugPrint()

	return C.LFS_ERR_OK
}

//export go_bdfs_erase_block
func go_bdfs_erase_block(flash_fs C.uintptr_t, block C.lfs_block_t) C.int {

	fs := (*FlashFS)(unsafe.Pointer(uintptr(flash_fs)))
	device_address := fs.AddressForBlock(uint32(block), 0)

	fs.device.EraseBlock(device_address)

	return C.LFS_ERR_OK
}
