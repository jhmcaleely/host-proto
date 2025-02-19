package main

/*
#include "lfs.h"
#include "bdfs_lfs_hal.h"
#include "block_device.h"
#include "pico_flash_fs.h"

int go_bdfs_read_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, void *buffer, lfs_size_t size);
int go_bdfs_erase_block_cgo(const struct lfs_config* c, lfs_block_t block);
int sync_block_nop(const struct lfs_config *c);

// configuration of the filesystem is provided by this struct
struct lfs_config cfg = {
    // block device operations
    .read  = go_bdfs_read_cgo,
    .prog  = bdfs_prog_page,
    .erase = go_bdfs_erase_block_cgo,
    .sync  = sync_block_nop,

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

struct block_device* bd;
int open_flags = LFS_O_RDWR | LFS_O_CREAT;

int go_bdfs_read(struct block_device* bd, uint32_t fs_base_address, lfs_block_t block, lfs_off_t off, void *buffer, lfs_size_t size);
int go_bdfs_erase_block(struct block_device* bd, uint32_t fs_base_address, lfs_block_t block);

int go_bdfs_read_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, void *buffer, lfs_size_t size) {

	struct flash_fs* fs = c->context;
    return go_bdfs_read(fs->device, fs->fs_flash_base_address, block, off, buffer, size);
}

int go_bdfs_erase_block_cgo(const struct lfs_config* c, lfs_block_t block) {

	struct flash_fs* fs = c->context;
    return go_bdfs_erase_block(fs->device, fs->fs_flash_base_address, block);
}


int sync_block_nop(const struct lfs_config *c) {

    return LFS_ERR_OK;
}
*/
import "C"
