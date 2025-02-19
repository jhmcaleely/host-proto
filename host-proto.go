package main

/*
#include "lfs.h"
#include "bdfs_lfs_hal.h"
#include "block_device.h"
#include "pico_flash_fs.h"

// configuration of the filesystem is provided by this struct
struct lfs_config cfg = {
    // block device operations
    .read  = bdfs_read,
    .prog  = bdfs_prog_page,
    .erase = bdfs_erase_block,
    .sync  = bdfs_sync_block,

    // block device configuration

    .read_size = 1,

    .prog_size = PICO_PROG_PAGE_SIZE,
    .block_size = PICO_ERASE_PAGE_SIZE,

    // the number of blocks we use for a flash fs.
    .block_count = FLASHFS_BLOCK_COUNT,

    // cache needs to be a multiple of the programming page size.
    .cache_size = PICO_PROG_PAGE_SIZE * 1,

    .lookahead_size = 16,
    .block_cycles = 500,
};

struct block_device* bd;

int _lfs_file_open(lfs_t* lfs, lfs_file_t *file, const char *path) {
	return lfs_file_open(lfs, file, path, LFS_O_RDWR | LFS_O_CREAT);
}
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

	C._lfs_file_open(fs, filep, C.CString("boot_count"))
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

func bdReadFromUF2(if2 io.Reader) {
	bd := C.bd

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
		if C.bdIsBlockStart(bd, C.uint32_t(ufn.TargetAddr)) {
			C.bdEraseBlock(bd, C.uint32_t(ufn.TargetAddr))
		}

		C.bdWrite(bd, C.uint32_t(ufn.TargetAddr), (*C.uint8_t)(unsafe.Pointer(&ufn.Data[0])), C.size_t(ufn.PayloadSize))
	}
}

func bdWriteToUF2(of io.Writer) {
	pageTotal := uint32(C.bdCountPages(C.bd))
	pageCursor := uint32(0)

	for b := uint32(0); b < PICO_DEVICE_BLOCK_COUNT; b++ {
		for p := uint32(0); p < PICO_FLASH_PAGE_PER_BLOCK; p++ {
			if C.bdPagePresent(C.bd, C.uint32_t(b), C.uint32_t(p)) {
				ub := Uf2Frame{}

				ub.MagicStart0 = UF2_MAGIC_START0
				ub.MagicStart1 = UF2_MAGIC_START1
				ub.Flags = UF2_FLAG_FAMILY_ID
				ub.TargetAddr = uint32(C.bdTargetAddress(C.bd, C.uint32_t(b), C.uint32_t(p)))
				ub.PayloadSize = PICO_PROG_PAGE_SIZE
				ub.BlockNo = pageCursor
				ub.NumBlocks = pageTotal

				// documented as FamilyID, Filesize or 0.
				ub.Reserved = PICO_UF2_FAMILYID

				C.bdRead(C.bd, C.uint32_t(ub.TargetAddr), (*C.uint8_t)(unsafe.Pointer(&ub.Data[0])), C.size_t(PICO_PROG_PAGE_SIZE))

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

	C.bd = C.bdCreate(C.uint32_t(PICO_FLASH_BASE_ADDR))
	defer C.bdDestroy(C.bd)

	C.bdfs_create_hal_at(&C.cfg, C.bd, C.uint32_t(FLASHFS_BASE_ADDR))
	defer C.bdfs_destroy_hal(&C.cfg)

	bdReadFromUF2(f)

	mount_and_update_boot()

	f.Seek(0, io.SeekStart)

	bdWriteToUF2(f)
}
