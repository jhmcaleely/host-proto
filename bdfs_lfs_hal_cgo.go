package main

/*
#include "lfs.h"

struct flash_fs;

int go_bdfs_read(struct flash_fs* fs, lfs_block_t block, lfs_off_t off, void *buffer, lfs_size_t size);
int go_bdfs_prog_page(struct flash_fs* fs, lfs_block_t block, lfs_off_t off, const void *buffer, lfs_size_t size);
int go_bdfs_erase_block(struct flash_fs* fs, lfs_block_t block);

int bdfs_read_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, void *buffer, lfs_size_t size) {

    return go_bdfs_read(c->context, block, off, buffer, size);
}

int bdfs_prog_page_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, const void *buffer, lfs_size_t size) {

    return go_bdfs_prog_page(c->context, block, off, buffer, size);
}

int bdfs_erase_block_cgo(const struct lfs_config* c, lfs_block_t block) {

    return go_bdfs_erase_block(c->context, block);
}
*/
import "C"
