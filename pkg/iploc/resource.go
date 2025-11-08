package iploc

import (
	"fmt"
	"io"
)

type resReadCloser interface {
	Read(b []byte) (n int, err error)
	ReadAt(b []byte, off int64) (n int, err error)
	Close() error
}

type resource struct {
	data []byte
	seek int64
}

func (res *resource) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}
	if max := len(res.data); len(b) > max {
		b = b[:max]
	}
	n, err = res.ReadAt(b, res.seek)
	res.seek += int64(n)
	return
}

func (res *resource) ReadAt(b []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, fmt.Errorf("negative offset: %d", off)
	} else if len(b) == 0 || off >= int64(len(res.data)) {
		return 0, nil
	}
	copy(b, res.data[off:])
	if int64(len(b))+off > int64(len(res.data)) {
		return int(int64(len(res.data)) - off), io.EOF
	}
	return len(b), nil
}

func (res *resource) Close() error {
	return nil
}
