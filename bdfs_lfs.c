#include "lfs.h"
#include "bdfs_lfs.h"
#include "pico_flash_device.h"

int bdfs_read_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, void *buffer, lfs_size_t size);
int bdfs_prog_page_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, const void *buffer, lfs_size_t size);
int bdfs_erase_block_cgo(const struct lfs_config* c, lfs_block_t block);
static int sync_block_nop(const struct lfs_config *c);

// configuration of the filesystem is provided by this function
void init_fscfg(struct lfs_config* cfg, struct flash_fs* fs, struct block_device* bd, uint32_t fs_base_address, uint32_t fs_block_count) {

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

    fs->device = bd;
    fs->fs_flash_base_address = fs_base_address;

    cfg->context = fs;
}

void destroy_fscfg(struct lfs_config* cfg, struct flash_fs* fs) {
    cfg->context = NULL;
}


static int sync_block_nop(const struct lfs_config *c) {

    return LFS_ERR_OK;
}