package main

import (
	"encoding/binary"
	"fmt"
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
	f, err := os.Open("test.uf2")
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
}
