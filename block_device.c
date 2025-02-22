#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <assert.h>
#include <stdbool.h>

#include "block_device.h"

#include "pico_flash_device.h"

/*
 * A RAM block device that mimics a Pico Flash device. We can write this 
 * out to a uf2 file for flashing.
 */

#define PICO_DEVICE_BLOCK_COUNT (PICO_FLASH_SIZE_BYTES / PICO_ERASE_PAGE_SIZE)
#define PICO_FLASH_PAGE_PER_BLOCK (PICO_ERASE_PAGE_SIZE / PICO_PROG_PAGE_SIZE)

struct block_device {
    uint8_t storage[PICO_FLASH_SIZE_BYTES];

    bool page_present[PICO_DEVICE_BLOCK_COUNT][PICO_FLASH_PAGE_PER_BLOCK];

    uint32_t base_address;
};

struct page_address {
    uint32_t block;
    uint32_t page;
    uint32_t offset;
};

uint32_t bdBaseAddress(struct block_device* bd) {
    return bd->base_address;
}

uint32_t bdStorageOffset(uint32_t block, uint32_t page) {
    return block * PICO_ERASE_PAGE_SIZE + page * PICO_PROG_PAGE_SIZE;
}

uint32_t getDeviceBlockNo(struct block_device* bd, uint32_t address) {
    return (address - bd->base_address) / PICO_ERASE_PAGE_SIZE;
}

void bdPageAddresss(struct block_device* bd, struct page_address* ad, uint32_t address) {
    uint32_t page_offset = (address - bd->base_address) % PICO_ERASE_PAGE_SIZE;    
    ad->page = page_offset / PICO_PROG_PAGE_SIZE;
    ad->offset = page_offset % PICO_PROG_PAGE_SIZE;
    ad->block = getDeviceBlockNo(bd, address);
}

struct block_device* bdCreate(uint32_t flash_base_address) {

    struct block_device* bd = malloc(sizeof(struct block_device));

    bd->base_address = flash_base_address;

    for (int b = 0; b < PICO_DEVICE_BLOCK_COUNT; b++) {
        for (int p = 0; p < PICO_FLASH_PAGE_PER_BLOCK; p++) {
            bd->page_present[b][p] = false;
        }
    }

    return bd;
}

void bdDestroy(struct block_device* bd) {
    free(bd);
}

void _bdEraseBlock(struct block_device* bd, uint32_t block) {

    for (int p = 0; p < PICO_FLASH_PAGE_PER_BLOCK; p++) {
        bd->page_present[block][p] = false;
    }
}

void _bdWrite(struct block_device* bd, uint32_t block, uint32_t page, const uint8_t* data, size_t size) {

    bd->page_present[block][page] = true;

    uint8_t* target = &bd->storage[bdStorageOffset(block, page)];

    memcpy(target, data, size);
}

void bdWrite(struct block_device* bd, uint32_t address, const uint8_t* data, size_t size) {
    assert(size <= PICO_PROG_PAGE_SIZE);

    struct page_address ad;
    bdPageAddresss(bd, &ad, address);

    assert(ad.offset == 0);

    _bdWrite(bd, ad.block, ad.page, data, size);
}

void _bdRead(struct block_device* bd, uint32_t block, uint32_t page, uint32_t off, uint8_t* buffer, size_t size) {

    uint32_t storage_offset = bdStorageOffset(block, page) + off;

    if (bd->page_present[block][page]) {
        printf("Read   available page");
        memcpy(buffer, &bd->storage[storage_offset], size);
    }
    else {
        printf("Read unavailable page");
    }
    printf("[%d][%d] off %d (size: %lu)\n", block, page, off, size);
}

void bdRead(struct block_device* bd, uint32_t address, uint8_t* buffer, size_t size) {

    struct page_address ad;
    bdPageAddresss(bd, &ad, address);

    _bdRead(bd, ad.block, ad.page, ad.offset, buffer, size);
}

int bdPagePresent(struct block_device* bd, uint32_t block, uint32_t page) {
    return bd->page_present[block][page];
}
