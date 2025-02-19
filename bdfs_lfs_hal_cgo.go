package main

/*
#include "lfs.h"
#include "pico_flash_device.h"

// Defines one region of flash to use for a filesystem. The size is a multiple of
// the 4096 byte erase size. We calculate it's location working back from the end of the
// flash device, so that code flashed at the start of the device will not collide.
// Pico's have a 2Mb flash device, so we're looking to be less than 2Mb.

// 128 blocks will reserve a 512K filsystem - 1/4 of the 2Mb device on a Pico

#define FLASHFS_BLOCK_COUNT 128
#define FLASHFS_SIZE_BYTES (PICO_ERASE_PAGE_SIZE * FLASHFS_BLOCK_COUNT)

// A start location counted back from the end of the device.
#define FLASHFS_BASE_ADDR (PICO_FLASH_BASE_ADDR + PICO_FLASH_SIZE_BYTES - FLASHFS_SIZE_BYTES)

int bdfs_read_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, void *buffer, lfs_size_t size);
int bdfs_prog_page_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, const void *buffer, lfs_size_t size);
int bdfs_erase_block_cgo(const struct lfs_config* c, lfs_block_t block);
int sync_block_nop(const struct lfs_config *c);

struct flash_fs {
    struct block_device* device;
    uint32_t fs_flash_base_address;
};

// configuration of the filesystem is provided by this function
void init_fscfg(struct lfs_config* cfg, struct block_device* bd, uint32_t fs_base_address, uint32_t fs_block_count) {

    cfg->read = bdfs_read_cgo;
    cfg->prog  = bdfs_prog_page_cgo;
    cfg->erase = bdfs_erase_block_cgo;
    cfg->sync  = sync_block_nop;

    // block device configuration
    cfg->read_size = 1;
    cfg->prog_size = PICO_PROG_PAGE_SIZE;
    cfg->block_size = PICO_ERASE_PAGE_SIZE;

    // the number of blocks we use for a flash fs.
    // Can be zero if we can read it from the fs.
    cfg->block_count = fs_block_count;

    // cache needs to be a multiple of the programming page size.
    cfg->cache_size = cfg->prog_size * 1;

    cfg->lookahead_size = 16;
    cfg->block_cycles = 500;

    struct flash_fs* fs = malloc(sizeof(struct flash_fs));
    fs->device = bd;
    fs->fs_flash_base_address = fs_base_address;

    cfg->context = fs;
}

void destroy_fscfg(struct lfs_config* cfg) {
    free(cfg->context);
    cfg->context = NULL;
}

int go_bdfs_read(struct block_device* bd, uint32_t fs_base_address, lfs_block_t block, lfs_off_t off, void *buffer, lfs_size_t size);
int go_bdfs_prog_page(struct block_device* bd, uint32_t fs_base_address, lfs_block_t block, lfs_off_t off, const void *buffer, lfs_size_t size);
int go_bdfs_erase_block(struct block_device* bd, uint32_t fs_base_address, lfs_block_t block);

int bdfs_read_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, void *buffer, lfs_size_t size) {

	struct flash_fs* fs = c->context;
    return go_bdfs_read(fs->device, fs->fs_flash_base_address, block, off, buffer, size);
}

int bdfs_prog_page_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, const void *buffer, lfs_size_t size) {

	struct flash_fs* fs = c->context;
	return go_bdfs_prog_page(fs->device, fs->fs_flash_base_address, block, off, buffer, size);
}

int bdfs_erase_block_cgo(const struct lfs_config* c, lfs_block_t block) {

	struct flash_fs* fs = c->context;
    return go_bdfs_erase_block(fs->device, fs->fs_flash_base_address, block);
}

int sync_block_nop(const struct lfs_config *c) {

    return LFS_ERR_OK;
}
*/
import "C"

const FLASHFS_BLOCK_COUNT = 128
const FLASHFS_SIZE_BYTES = PICO_ERASE_PAGE_SIZE * FLASHFS_BLOCK_COUNT

// A start location counted back from the end of the device.
const FLASHFS_BASE_ADDR uint32 = PICO_FLASH_BASE_ADDR + PICO_FLASH_SIZE_BYTES - FLASHFS_SIZE_BYTES
