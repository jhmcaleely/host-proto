#ifndef block_device_h
#define block_device_h

#include <stdint.h>
#include <stdbool.h>

struct block_device;

struct block_device* bdCreate(uint32_t flash_base_address);
void bdDestroy(struct block_device* bd);

void _bdEraseBlock(struct block_device* bd, uint32_t block);

void bdWrite(struct block_device* bd, uint32_t address, const uint8_t* data, size_t size);
void bdRead(struct block_device* bd, uint32_t address, uint8_t* buffer, size_t size);

int bdPagePresent(struct block_device* bd, uint32_t block, uint32_t page);
uint32_t bdBaseAddress(struct block_device* bd);

#endif