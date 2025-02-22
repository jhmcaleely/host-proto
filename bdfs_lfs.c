#include "lfs.h"
#include "bdfs_lfs.h"

int bdfs_read_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, void *buffer, lfs_size_t size);
int bdfs_prog_page_cgo(const struct lfs_config* c, lfs_block_t block, lfs_off_t off, const void *buffer, lfs_size_t size);
int bdfs_erase_block_cgo(const struct lfs_config* c, lfs_block_t block);

static int sync_nop(const struct lfs_config *c) {

    return LFS_ERR_OK;
}

void install_bdfs_hooks(struct lfs_config* cfg, struct flash_fs* fs) {
    cfg->read = bdfs_read_cgo;
    cfg->prog  = bdfs_prog_page_cgo;
    cfg->erase = bdfs_erase_block_cgo;
    cfg->sync  = sync_nop;

    cfg->context = fs;
}

void remove_bdfs_hooks(struct lfs_config* cfg, struct flash_fs* fs) {
    cfg->read = NULL;
    cfg->prog  = NULL;
    cfg->erase = NULL;
    cfg->sync  = NULL;

    cfg->context = NULL;
}
