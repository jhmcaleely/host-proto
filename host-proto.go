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

void _bdInit() {
   bd = bdCreate(PICO_FLASH_BASE_ADDR);
   bdfs_create_hal_at(&cfg, bd, FLASHFS_BASE_ADDR);
}

void _bdDestroy() {
    bdfs_destroy_hal(&cfg);
    bdDestroy(bd);
}

int _lfs_file_open(lfs_t* lfs, lfs_file_t *file, const char *path) {
	return lfs_file_open(lfs, file, path, LFS_O_RDWR | LFS_O_CREAT);
}

void writeuf2(const char * input) {
    FILE* iofile = fopen(input, "wb");
    if (iofile) {
        bdWriteToUF2(bd, iofile);

        fclose(iofile);
    }
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

const UF2_MAGIC_START0 = 0x0A324655
const UF2_MAGIC_START1 = 0x9E5D5157
const UF2_MAGIC_END = 0x0AB16F30

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

func main() {
	f, err := os.Open("test.uf2")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	C._bdInit()
	defer C._bdDestroy()

	bdReadFromUF2(f)

	mount_and_update_boot()

	C.writeuf2(C.CString("test.uf2"))

}
