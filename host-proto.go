package main

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

	uf := Uf2Frame{}
	err = binary.Read(f, binary.LittleEndian, &uf)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Print(uf.NumBlocks)
	fmt.Print(uf.BlockNo)
	fmt.Print(uf.MagicStart0)
	fmt.Print(uf.Flags)

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
