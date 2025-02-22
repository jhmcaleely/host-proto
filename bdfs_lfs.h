#ifndef bdfs_lfs_h
#define bdfs_lfs_h

#include <stdint.h>

// Defines one region of flash to use for a filesystem. The size is a multiple of
// the 4096 byte erase size. We calculate it's location working back from the end of the
// flash device, so that code flashed at the start of the device will not collide.
// Pico's have a 2Mb flash device, so we're looking to be less than 2Mb.

// 128 blocks will reserve a 512K filsystem - 1/4 of the 2Mb device on a Pico

#define FLASHFS_BLOCK_COUNT 128
#define FLASHFS_SIZE_BYTES (PICO_ERASE_PAGE_SIZE * FLASHFS_BLOCK_COUNT)

// A start location counted back from the end of the device.
#define FLASHFS_BASE_ADDR (PICO_FLASH_BASE_ADDR + PICO_FLASH_SIZE_BYTES - FLASHFS_SIZE_BYTES)

struct lfs_config;
struct block_device;

struct flash_fs {
    struct block_device* device;
    uint32_t fs_flash_base_address;
};

void install_bdfs_hooks(struct lfs_config* cfg, struct flash_fs* fs);
void remove_bdfs_hooks(struct lfs_config* cfg, struct flash_fs* fs);

#endif