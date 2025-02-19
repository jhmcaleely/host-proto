package main

/*
#include "lfs.h"
#include "block_device.h"

void bdfs_create_hal_at(struct lfs_config* c, struct block_device* bd, uint32_t fs_base_address);
void bdfs_destroy_hal(struct lfs_config* c);

extern int open_flags;
extern struct lfs_config cfg;

*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"unsafe"
)

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

type Uf2Frame struct {
	MagicStart0 uint32
	MagicStart1 uint32
	Flags       uint32
	TargetAddr  uint32
	PayloadSize uint32
	BlockNo     uint32
	NumBlocks   uint32
	Reserved    uint32
	Data        [476]byte
	MagicEnd    uint32
}

const UF2_MAGIC_START0 uint32 = 0x0A324655
const UF2_MAGIC_START1 uint32 = 0x9E5D5157
const UF2_MAGIC_END uint32 = 0x0AB16F30
const UF2_FLAG_NOFLASH uint32 = 0x00000001
const UF2_FLAG_FILECONTAINER uint32 = 0x00001000
const UF2_FLAG_FAMILY_ID uint32 = 0x00002000
const UF2_FLAG_MD5_CHKSUM uint32 = 0x00004000
const UF2_FLAG_EXTENSION_TAGS uint32 = 0x00008000

const PICO_FLASH_BASE_ADDR uint32 = 0x10000000
const PICO_FLASH_SIZE_BYTES = (2 * 1024 * 1024)
const PICO_ERASE_PAGE_SIZE = 4096
const PICO_PROG_PAGE_SIZE = 256

const PICO_DEVICE_BLOCK_COUNT = PICO_FLASH_SIZE_BYTES / PICO_ERASE_PAGE_SIZE
const PICO_FLASH_PAGE_PER_BLOCK = PICO_ERASE_PAGE_SIZE / PICO_PROG_PAGE_SIZE

const PICO_UF2_FAMILYID uint32 = 0xe48bff56

const FLASHFS_BLOCK_COUNT = 128
const FLASHFS_SIZE_BYTES = PICO_ERASE_PAGE_SIZE * FLASHFS_BLOCK_COUNT

// A start location counted back from the end of the device.
const FLASHFS_BASE_ADDR uint32 = PICO_FLASH_BASE_ADDR + PICO_FLASH_SIZE_BYTES - FLASHFS_SIZE_BYTES

func update_boot_count(fs *C.lfs_t) {
	var lfsfile C.lfs_file_t
	var pin runtime.Pinner
	defer pin.Unpin()

	filep := &lfsfile
	pin.Pin(filep)

	C.lfs_file_open(fs, filep, C.CString("boot_count"), C.open_flags)
	defer C.lfs_file_close(fs, filep)

	var boot_count C.uint32_t
	C.lfs_file_read(fs, filep, unsafe.Pointer(&boot_count), 4)
	// update boot count
	boot_count += 1
	C.lfs_file_rewind(fs, filep)

	C.lfs_file_write(fs, filep, unsafe.Pointer(&boot_count), 4)

	fmt.Printf("boot count: %d\n", boot_count)
}

func mount_and_update_boot() {
	var lfs C.lfs_t
	var pin runtime.Pinner
	defer pin.Unpin()

	lfsp := &lfs
	pin.Pin(lfsp)

	cfgp := &C.cfg

	lfsres := C.lfs_mount(lfsp, cfgp)
	if lfsres != 0 {
		C.lfs_format(lfsp, cfgp)
		C.lfs_mount(lfsp, cfgp)
	}
	defer C.lfs_unmount(lfsp)

	update_boot_count(lfsp)
}

type BdFS struct {
	Device      *C.struct_block_device
	BaseAddress uint32
}

func bdReadFromUF2(device BdFS, if2 io.Reader) {

	ufn := Uf2Frame{}
	for binary.Read(if2, binary.LittleEndian, &ufn) != io.EOF {

		if ufn.MagicStart0 != UF2_MAGIC_START0 {
			panic("bad start0")
		}
		if ufn.MagicStart1 != UF2_MAGIC_START1 {
			panic("bad start1")
		}
		if ufn.MagicEnd != UF2_MAGIC_END {
			panic("bad end")
		}

		// erase a block before writing any pages to it.
		if C.bdIsBlockStart(device.Device, C.uint32_t(ufn.TargetAddr)) {
			C.bdEraseBlock(device.Device, C.uint32_t(ufn.TargetAddr))
		}

		C.bdWrite(device.Device, C.uint32_t(ufn.TargetAddr), (*C.uint8_t)(unsafe.Pointer(&ufn.Data[0])), C.size_t(ufn.PayloadSize))
	}
}

func bdWriteToUF2(device BdFS, of io.Writer) {
	pageTotal := uint32(C.bdCountPages(device.Device))
	pageCursor := uint32(0)

	for b := uint32(0); b < PICO_DEVICE_BLOCK_COUNT; b++ {
		for p := uint32(0); p < PICO_FLASH_PAGE_PER_BLOCK; p++ {
			if C.bdPagePresent(device.Device, C.uint32_t(b), C.uint32_t(p)) {
				ub := Uf2Frame{}

				ub.MagicStart0 = UF2_MAGIC_START0
				ub.MagicStart1 = UF2_MAGIC_START1
				ub.Flags = UF2_FLAG_FAMILY_ID
				ub.TargetAddr = uint32(C.bdTargetAddress(device.Device, C.uint32_t(b), C.uint32_t(p)))
				ub.PayloadSize = PICO_PROG_PAGE_SIZE
				ub.BlockNo = pageCursor
				ub.NumBlocks = pageTotal

				// documented as FamilyID, Filesize or 0.
				ub.Reserved = PICO_UF2_FAMILYID

				C.bdRead(device.Device, C.uint32_t(ub.TargetAddr), (*C.uint8_t)(unsafe.Pointer(&ub.Data[0])), C.size_t(PICO_PROG_PAGE_SIZE))

				ub.MagicEnd = UF2_MAGIC_END

				fmt.Printf("uf2page: %08x, %d\n", ub.TargetAddr, ub.PayloadSize)

				binary.Write(of, binary.LittleEndian, &ub)

				pageCursor++
			}
		}
	}
}

func main() {
	f, err := os.OpenFile("test.uf2", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	fs := BdFS{}
	fs.BaseAddress = FLASHFS_BASE_ADDR
	fs.Device = C.bdCreate(C.uint32_t(PICO_FLASH_BASE_ADDR))
	defer C.bdDestroy(fs.Device)

	C.bdfs_create_hal_at(&C.cfg, fs.Device, C.uint32_t(FLASHFS_BASE_ADDR))
	defer C.bdfs_destroy_hal(&C.cfg)

	bdReadFromUF2(fs, f)

	mount_and_update_boot()

	f.Seek(0, io.SeekStart)

	bdWriteToUF2(fs, f)
}
