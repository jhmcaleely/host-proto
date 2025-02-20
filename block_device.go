package main

/*
#include "block_device.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"io"
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

const PICO_DEVICE_BLOCK_COUNT = PICO_FLASH_SIZE_BYTES / PICO_ERASE_PAGE_SIZE
const PICO_FLASH_PAGE_PER_BLOCK = PICO_ERASE_PAGE_SIZE / PICO_PROG_PAGE_SIZE

func bdReadFromUF2(device *C.struct_block_device, if2 io.Reader) {

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
		if C.bdIsBlockStart(device, C.uint32_t(ufn.TargetAddr)) {
			C.bdEraseBlock(device, C.uint32_t(ufn.TargetAddr))
		}

		C.bdWrite(device, C.uint32_t(ufn.TargetAddr), (*C.uint8_t)(unsafe.Pointer(&ufn.Data[0])), C.size_t(ufn.PayloadSize))
	}
}

func bdWriteToUF2(device *C.struct_block_device, of io.Writer) {
	pageTotal := uint32(C.bdCountPages(device))
	pageCursor := uint32(0)

	for b := uint32(0); b < PICO_DEVICE_BLOCK_COUNT; b++ {
		for p := uint32(0); p < PICO_FLASH_PAGE_PER_BLOCK; p++ {
			if C.bdPagePresent(device, C.uint32_t(b), C.uint32_t(p)) {
				ub := Uf2Frame{}

				ub.MagicStart0 = UF2_MAGIC_START0
				ub.MagicStart1 = UF2_MAGIC_START1
				ub.Flags = UF2_FLAG_FAMILY_ID
				ub.TargetAddr = uint32(C.bdTargetAddress(device, C.uint32_t(b), C.uint32_t(p)))
				ub.PayloadSize = PICO_PROG_PAGE_SIZE
				ub.BlockNo = pageCursor
				ub.NumBlocks = pageTotal

				// documented as FamilyID, Filesize or 0.
				ub.Reserved = PICO_UF2_FAMILYID

				C.bdRead(device, C.uint32_t(ub.TargetAddr), (*C.uint8_t)(unsafe.Pointer(&ub.Data[0])), C.size_t(PICO_PROG_PAGE_SIZE))

				ub.MagicEnd = UF2_MAGIC_END

				fmt.Printf("uf2page: %08x, %d\n", ub.TargetAddr, ub.PayloadSize)

				binary.Write(of, binary.LittleEndian, &ub)

				pageCursor++
			}
		}
	}
}
