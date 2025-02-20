package main

/*
#include "lfs.h"
#include "bdfs_lfs.h"
#include "block_device.h"

int open_flags = LFS_O_RDWR | LFS_O_CREAT;
*/
import "C"

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"unsafe"
)

const PICO_FLASH_BASE_ADDR uint32 = 0x10000000
const PICO_FLASH_SIZE_BYTES = (2 * 1024 * 1024)
const PICO_ERASE_PAGE_SIZE = 4096
const PICO_PROG_PAGE_SIZE = 256

const PICO_UF2_FAMILYID uint32 = 0xe48bff56

type BdFS struct {
	LfsConfig C.struct_lfs_config
	FsConfig  C.struct_flash_fs

	LfsP *C.struct_lfs_config
	FsP  *C.struct_flash_fs

	Device *C.struct_block_device
}

func newBdFS(device *C.struct_block_device) *BdFS {
	cfg := BdFS{}
	cfg.LfsP = &cfg.LfsConfig
	cfg.FsP = &cfg.FsConfig
	cfg.Device = device

	return &cfg
}

func (fs *BdFS) init(baseAddr, blockCount uint32) error {

	C.init_fscfg(fs.LfsP, fs.FsP, fs.Device, C.uint32_t(baseAddr), C.uint32_t(blockCount))

	return nil
}

func (fs *BdFS) Close() error {
	C.destroy_fscfg(fs.LfsP, fs.FsP)
	return nil
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

	lfsres := C.lfs_mount(lfsp, fs.LfsP)
	if lfsres != 0 {
		C.lfs_format(lfsp, fs.LfsP)
		C.lfs_mount(lfsp, fs.LfsP)
	}
	defer C.lfs_unmount(lfsp)

	update_boot_count(lfsp)
}

func bootCountDemo(fsFilename string) {
	f, err := os.OpenFile(fsFilename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	device := C.bdCreate(C.uint32_t(PICO_FLASH_BASE_ADDR))
	defer C.bdDestroy(device)

	fs := newBdFS(device)
	var pin runtime.Pinner
	pin.Pin(fs.LfsP)
	defer pin.Unpin()
	pin.Pin(fs.FsP)

	defer fs.Close()

	fs.init(FLASHFS_BASE_ADDR, FLASHFS_BLOCK_COUNT)

	bdReadFromUF2(device, f)

	mount_and_update_boot(fs)

	f.Seek(0, io.SeekStart)

	bdWriteToUF2(device, f)
}

func main() {

	bootCountDemoCmd := flag.NewFlagSet("bootcount", flag.ExitOnError)
	bootCountFS := bootCountDemoCmd.String("fs", "test.uf2", "mount and increment boot_count on fs")

	if len(os.Args) < 2 {
		fmt.Println("expected command")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "bootcount":
		bootCountDemoCmd.Parse(os.Args[2:])
		bootCountDemo(*bootCountFS)
	default:
		fmt.Println("unknown command")
		os.Exit(1)
	}
}
