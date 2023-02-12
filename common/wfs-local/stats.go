// +build linux darwin

package local

import (
	"syscall"
)

func getFSSize(root string) (uint64, uint64, error) {
	var stat syscall.Statfs_t

	err := syscall.Statfs(root, &stat)
	if err != nil {
		return 0, 0, err
	}

	free := stat.Bavail * uint64(stat.Bsize)
	total := stat.Blocks * uint64(stat.Bsize)
	used := total - free

	return used, free, nil
}
