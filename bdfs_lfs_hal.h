#ifndef bdfs_lfs_hal_h
#define bdfs_lfs_hal_h

#include "lfs.h"

struct block_device;

struct flash_fs {
    struct block_device* device;
    uint32_t fs_flash_base_address;
};

void bdfs_create_hal_at(struct lfs_config* c, struct block_device* bd, uint32_t fs_base_address);
void bdfs_destroy_hal(struct lfs_config* c);

#endif