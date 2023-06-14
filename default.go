// Copyright 2022 someonegg. All rights reserscoreed.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nodefs

import (
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
)

// DefaultFileSystem returns a filesystem that responds ENOSYS for
// all methods
func DefaultFileSystem() FileSystem {
	return defaultFileSystem{}
}

type defaultFileSystem struct{}

// uFh may be 0.
func (fs defaultFileSystem) GetAttr(ctx *Context, ino uint64, uFh uint32, out *fuse.Attr) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Access(ctx *Context, ino uint64, mask uint32) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Lookup(ctx *Context, parentIno uint64, name string, out *fuse.EntryOut) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Forget(ino uint64, nlookup uint64) fuse.Status {
	return fuse.ENOSYS
}

// Tree structure
func (fs defaultFileSystem) Mknod(ctx *Context, parentIno uint64, name string, mode uint32, dev uint32) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Mkdir(ctx *Context, parentIno uint64, name string, mode uint32) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Unlink(ctx *Context, parentIno uint64, name string) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Rmdir(ctx *Context, parentIno uint64, name string) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Rename(ctx *Context, parentIno uint64, name string, newParentIno uint64, newName string) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Link(ctx *Context, ino uint64, newParentIno uint64, newName string) fuse.Status {
	return fuse.ENOSYS
}

// Symlinks
func (fs defaultFileSystem) Symlink(ctx *Context, parentIno uint64, name string, target string) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Readlink(ctx *Context, ino uint64) (target string, code fuse.Status) {
	return "", fuse.ENOSYS
}

// Extended attributes
func (fs defaultFileSystem) GetXAttr(ctx *Context, ino uint64, attr string) (data []byte, code fuse.Status) {
	return nil, fuse.ENOSYS
}

func (fs defaultFileSystem) ListXAttr(ctx *Context, ino uint64) (attrs []string, code fuse.Status) {
	return nil, fuse.ENOSYS
}

func (fs defaultFileSystem) SetXAttr(ctx *Context, ino uint64, attr string, data []byte, flags uint32) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) RemoveXAttr(ctx *Context, ino uint64, attr string) fuse.Status {
	return fuse.ENOSYS
}

// File
func (fs defaultFileSystem) Create(ctx *Context, parentIno uint64, name string, flags uint32, mode uint32) (uFh uint32, forceDIO bool, code fuse.Status) {
	return 0, false, fuse.ENOSYS
}

func (fs defaultFileSystem) Open(ctx *Context, ino uint64, flags uint32) (uFh uint32, keepCache, forceDIO bool, code fuse.Status) {
	return 0, false, false, fuse.ENOSYS
}

func (fs defaultFileSystem) Read(ctx *Context, ino uint64, uFh uint32, dest []byte, off uint64) (result fuse.ReadResult, code fuse.Status) {
	return nil, fuse.ENOSYS
}

func (fs defaultFileSystem) Write(ctx *Context, ino uint64, uFh uint32, data []byte, off uint64) (written uint32, code fuse.Status) {
	return 0, fuse.ENOSYS
}

func (fs defaultFileSystem) Fallocate(ctx *Context, ino uint64, uFh uint32, off uint64, size uint64, mode uint32) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Fsync(ctx *Context, ino uint64, uFh uint32, flags uint32) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Flush(ctx *Context, ino uint64, uFh uint32) fuse.Status {
	return fuse.ENOSYS
}
func (fs defaultFileSystem) Release(ctx *Context, ino uint64, uFh uint32) {}

func (fs defaultFileSystem) GetLk(ctx *Context, ino uint64, uFh uint32, owner uint64, lk *fuse.FileLock, flags uint32, out *fuse.FileLock) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) SetLk(ctx *Context, ino uint64, uFh uint32, owner uint64, lk *fuse.FileLock, flags uint32) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) SetLkw(ctx *Context, ino uint64, uFh uint32, owner uint64, lk *fuse.FileLock, flags uint32) fuse.Status {
	return fuse.ENOSYS
}

// uFh may be 0.
func (fs defaultFileSystem) Chmod(ctx *Context, ino uint64, uFh uint32, mode uint32) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Chown(ctx *Context, ino uint64, uFh uint32, uid uint32, gid uint32) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Truncate(ctx *Context, ino uint64, uFh uint32, size uint64) fuse.Status {
	return fuse.ENOSYS
}

func (fs defaultFileSystem) Utimens(ctx *Context, ino uint64, uFh uint32, atime *time.Time, mtime *time.Time) fuse.Status {
	return fuse.ENOSYS
}

// Directory
func (fs defaultFileSystem) Lsdir(ctx *Context, ino uint64) (stream []fuse.DirEntry, code fuse.Status) {
	return nil, fuse.ENOSYS
}

func (fs defaultFileSystem) StatFs(ctx *Context, ino uint64, out *fuse.StatfsOut) fuse.Status {
	return fuse.OK
}
