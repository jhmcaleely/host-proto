#include "lfs.h"
#include "bdfs_lfs.h"
#include "pico_flash_device.h"

int bdfs_read_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, void *buffer, lfs_size_t size);
int bdfs_prog_page_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, const void *buffer, lfs_size_t size);
int bdfs_erase_block_cgo(const struct lfs_config* c, lfs_block_t block);
static int sync_block_nop(const struct lfs_config *c);

// configuration of the filesystem is provided by this function
void install_bdfs_hooks(struct lfs_config* cfg, struct flash_fs* fs, struct block_device* bd, uint32_t fs_base_address) {

    cfg->read = bdfs_read_cgo;
    cfg->prog  = bdfs_prog_page_cgo;
    cfg->erase = bdfs_erase_block_cgo;
    cfg->sync  = sync_block_nop;

    fs->device = bd;
    fs->fs_flash_base_address = fs_base_address;

    cfg->context = fs;
}

void remove_bdfs_hooks(struct lfs_config* cfg, struct flash_fs* fs) {
    cfg->read = NULL;
    cfg->prog  = NULL;
    cfg->erase = NULL;
    cfg->sync  = NULL;

    cfg->context = NULL;
}

static int sync_block_nop(const struct lfs_config *c) {

    return LFS_ERR_OK;
}