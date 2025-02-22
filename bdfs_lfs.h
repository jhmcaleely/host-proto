#ifndef bdfs_lfs_h
#define bdfs_lfs_h

#include <stdint.h>

struct lfs_config;
struct block_device;

struct flash_fs {
    struct block_device* device;
    uint32_t fs_flash_base_address;
};

void install_bdfs_hooks(struct lfs_config* cfg, struct flash_fs* fs);
void remove_bdfs_hooks(struct lfs_config* cfg, struct flash_fs* fs);

#endif