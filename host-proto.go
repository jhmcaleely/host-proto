package main

/*
#include "lfs.h"
#include "bdfs_lfs.h"
#include "block_device.h"
*/
import "C"

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
)

const PICO_FLASH_BASE_ADDR uint32 = 0x10000000
const PICO_FLASH_SIZE_BYTES = (2 * 1024 * 1024)
const PICO_ERASE_PAGE_SIZE = 4096
const PICO_PROG_PAGE_SIZE = 256

const PICO_UF2_FAMILYID uint32 = 0xe48bff56

type BdFS struct {
	cfg *LittleFsConfig
	FsP *C.struct_flash_fs

	Device *C.struct_block_device
}

func newBdFS(device *C.struct_block_device, blockCount uint32) *BdFS {
	cfg := BdFS{cfg: newLittleFsConfig(blockCount)}

	var blockfs C.struct_flash_fs
	cfg.FsP = &blockfs

	cfg.Device = device

	return &cfg
}

func (fs *BdFS) init(baseAddr uint32) error {

	C.install_bdfs_hooks(fs.cfg.chandle, fs.FsP, fs.Device, C.uint32_t(baseAddr))

	return nil
}

func (fs *BdFS) Close() error {
	C.remove_bdfs_hooks(fs.cfg.chandle, fs.FsP)
	return nil
}

func (fs *BdFS) ensure_mount() *LittleFs {
	var lfs *LittleFs

	lfs, err := fs.cfg.Mount()
	if err != nil {
		lfsFormat(fs.cfg.chandle)
		lfs, err = fs.cfg.Mount()
	}
	return lfs
}

func update_boot_count(lfs *LittleFs) {

	file := newLfsFile(lfs)
	file.Open("boot_count")
	defer file.Close()

	var boot_count uint32
	binary.Read(file, binary.LittleEndian, &boot_count)

	boot_count += 1
	file.Rewind()

	binary.Write(file, binary.LittleEndian, boot_count)

	fmt.Printf("boot count: %d\n", boot_count)
}

func add_file(lfs *LittleFs, fileToAdd string) {

	r, err := os.Open(fileToAdd)
	if err != nil {
		fmt.Println(("nothing to open"))
		os.Exit(1)
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		fmt.Println("nothing read")
		os.Exit(1)
	}

	file := newLfsFile(lfs)
	file.Open(fileToAdd)
	defer file.Close()

	file.Write(data)
}

func bootCountDemo(fsFilename string) {
	f, err := os.OpenFile(fsFilename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	device := C.bdCreate(C.uint32_t(PICO_FLASH_BASE_ADDR))
	defer C.bdDestroy(device)

	fs := newBdFS(device, FLASHFS_BLOCK_COUNT)
	var pin runtime.Pinner
	pin.Pin(fs.cfg.chandle)
	defer pin.Unpin()
	pin.Pin(fs.FsP)

	defer fs.Close()

	fs.init(FLASHFS_BASE_ADDR)

	bdReadFromUF2(device, f)

	lfs := fs.ensure_mount()
	defer lfs.Close()

	update_boot_count(lfs)

	f.Seek(0, io.SeekStart)

	bdWriteToUF2(device, f)
}

func addFile(fsFilename, fileToAdd string) {
	f, err := os.OpenFile(fsFilename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	device := C.bdCreate(C.uint32_t(PICO_FLASH_BASE_ADDR))
	defer C.bdDestroy(device)

	fs := newBdFS(device, FLASHFS_BLOCK_COUNT)
	var pin runtime.Pinner
	pin.Pin(fs.cfg.chandle)
	defer pin.Unpin()
	pin.Pin(fs.FsP)

	defer fs.Close()

	fs.init(FLASHFS_BASE_ADDR)

	bdReadFromUF2(device, f)

	lfs := fs.ensure_mount()
	defer lfs.Close()

	add_file(lfs, fileToAdd)

	f.Seek(0, io.SeekStart)

	bdWriteToUF2(device, f)
}

func list_files(fs *LittleFs, dirEntry string) {

	dir := LfsDir{Lfs: fs}
	err := dir.Open(dirEntry)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(10)
	}
	defer dir.Close()

	var info C.struct_lfs_info

	for more, err := dir.Read(&info); more; more, err = dir.Read(&info) {
		if err != nil {
			os.Exit(10)
		}
		fmt.Println(C.GoString(&info.name[0]))
	}

}

func lsDir(fsFilename, dirEntry string) {
	f, err := os.OpenFile(fsFilename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	device := C.bdCreate(C.uint32_t(PICO_FLASH_BASE_ADDR))
	defer C.bdDestroy(device)

	fs := newBdFS(device, FLASHFS_BLOCK_COUNT)
	var pin runtime.Pinner
	pin.Pin(fs.cfg.chandle)
	defer pin.Unpin()
	pin.Pin(fs.FsP)

	defer fs.Close()

	fs.init(FLASHFS_BASE_ADDR)

	bdReadFromUF2(device, f)

	lfs := fs.ensure_mount()
	defer lfs.Close()

	list_files(lfs, dirEntry)
}

func main() {

	bootCountDemoCmd := flag.NewFlagSet("bootcount", flag.ExitOnError)
	bootCountFS := bootCountDemoCmd.String("fs", "test.uf2", "mount and increment boot_count on fs")

	addFileCmd := flag.NewFlagSet("addfile", flag.ExitOnError)
	addFileFS := addFileCmd.String("fs", "test.uf2", "add file to this filesystem")
	addFileName := addFileCmd.String("add", "", "filename to add")

	lsDirCmd := flag.NewFlagSet("ls", flag.ExitOnError)
	lsDirFS := lsDirCmd.String("fs", "test.uf2", "filesystem to mount")
	lsDirEntry := lsDirCmd.String("dir", "/", "directory to ls")

	if len(os.Args) < 2 {
		fmt.Println("expected command")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "bootcount":
		bootCountDemoCmd.Parse(os.Args[2:])
		bootCountDemo(*bootCountFS)
	case "addfile":
		addFileCmd.Parse(os.Args[2:])
		if *addFileName == "" {
			fmt.Println("expect filename to add")
			os.Exit(1)
		}
		addFile(*addFileFS, *addFileName)
	case "ls":
		lsDirCmd.Parse((os.Args[2:]))
		lsDir(*lsDirFS, *lsDirEntry)
	default:
		fmt.Println("unknown command")
		os.Exit(1)
	}
}
