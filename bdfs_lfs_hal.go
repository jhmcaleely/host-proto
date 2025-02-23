package main

type FlashFS struct {
	device       BlockDevice
	base_address uint32
}

func (fs FlashFS) AddressForBlock(block uint32, off uint32) uint32 {

	byte_offset := block*PICO_ERASE_PAGE_SIZE + off

	return fs.base_address + byte_offset
}
