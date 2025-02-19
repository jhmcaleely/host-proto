package main

/*
#include "lfs.h"
#include "bdfs_lfs.h"
#include "block_device.h"

void init_fscfg(struct lfs_config* cfg, struct block_device* bd, uint32_t fs_base_address, uint32_t fs_block_count);
void destroy_fscfg(struct lfs_config* cfg);

int open_flags = LFS_O_RDWR | LFS_O_CREAT;
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

type BdFS struct {
	FsConfig    C.struct_lfs_config
	Device      *C.struct_block_device
	BaseAddress uint32
	BlockCount  uint32
}

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

func mount_and_update_boot(fs *BdFS) {
	var lfs C.lfs_t
	var pin runtime.Pinner
	defer pin.Unpin()

	lfsp := &lfs
	pin.Pin(lfsp)

	lfsres := C.lfs_mount(lfsp, &fs.FsConfig)
	if lfsres != 0 {
		C.lfs_format(lfsp, &fs.FsConfig)
		C.lfs_mount(lfsp, &fs.FsConfig)
	}
	defer C.lfs_unmount(lfsp)

	update_boot_count(lfsp)
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
	var pin runtime.Pinner
	fsp := &fs.FsConfig
	pin.Pin(fsp)
	defer pin.Unpin()

	fs.Device = C.bdCreate(C.uint32_t(PICO_FLASH_BASE_ADDR))
	defer C.bdDestroy(fs.Device)
	fs.BaseAddress = FLASHFS_BASE_ADDR
	fs.BlockCount = FLASHFS_BLOCK_COUNT

	C.init_fscfg(fsp, fs.Device, C.uint32_t(fs.BaseAddress), C.uint32_t(fs.BlockCount))
	defer C.destroy_fscfg(fsp)

	bdReadFromUF2(fs, f)

	mount_and_update_boot(&fs)

	f.Seek(0, io.SeekStart)

	bdWriteToUF2(fs, f)
}
