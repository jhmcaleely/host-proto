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

lfs_t lfs;
lfs_file_t file;

struct block_device* bd;

void _bdInit() {
   bd = bdCreate(PICO_FLASH_BASE_ADDR);
   bdfs_create_hal_at(&cfg, bd, FLASHFS_BASE_ADDR);
}

int _lfs_mount() {
	return lfs_mount(&lfs, &cfg);
}

int _lfs_format() {
	return lfs_format(&lfs, &cfg);
}
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
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

func main() {
	f, err := os.Open("littlefs-pico.uf2")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	C._bdInit()
	lfsres := C._lfs_mount()
	if lfsres != 0 {
		C._lfs_format()
		lfsres = C._lfs_mount()
	}
	fmt.Print(lfsres)

	uf := Uf2Frame{}
	err = binary.Read(f, binary.LittleEndian, &uf)
	if err != nil {
		log.Fatalln(err)
	}

	if2, err := os.Open("fs.uf2")
	if err != nil {
		log.Fatalln(err)
	}
	defer if2.Close()

	ufCount := uf.NumBlocks

	ufn := Uf2Frame{}
	err = binary.Read(if2, binary.LittleEndian, &ufn)
	if err != nil {
		log.Fatalln(err)
	}

	ufnCount := ufn.NumBlocks

	ofCount := ufCount + ufnCount

	if2.Seek(0, io.SeekStart)
	f.Seek(0, io.SeekStart)

	of, err := os.Create("testout.uf2")
	if err != nil {
		log.Fatalln(err)
	}
	defer of.Close()

	for u := range ufCount {
		binary.Read(f, binary.LittleEndian, &uf)
		uf.BlockNo = u
		uf.NumBlocks = ofCount
		binary.Write(of, binary.LittleEndian, &uf)
	}

	for u := ufCount; u < ofCount; u++ {
		binary.Read(if2, binary.LittleEndian, &uf)
		uf.BlockNo = u
		uf.NumBlocks = ofCount
		binary.Write(of, binary.LittleEndian, &uf)
	}

}
