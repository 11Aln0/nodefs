package nodefs

import (
	"github.com/hanwen/go-fuse/v2/fuse"
	"log"
	"time"
)

// FileSystem API that uses ino to operate files. A minimal
// file system should have at least a functional GetAttr method, and
// the returned attr needs to have a valid Ino.
// Typically, each call happens in its own goroutine, so take care to
// make the file system thread-safe.
type FileSystem interface {
	// uFh may be 0.
	GetAttr(ctx *Context, ino uint64, uFh uint32, out *fuse.Attr) fuse.Status

	Access(ctx *Context, ino uint64, mask uint32) fuse.Status

	Lookup(ctx *Context, parentIno uint64, name string, out *fuse.Attr) fuse.Status

	// Tree structure
	Mknod(ctx *Context, parentIno uint64, name string, mode uint32, dev uint32) fuse.Status
	Mkdir(ctx *Context, parentIno uint64, name string, mode uint32) fuse.Status
	Unlink(ctx *Context, parentIno uint64, name string) fuse.Status
	Rmdir(ctx *Context, parentIno uint64, name string) fuse.Status
	Rename(ctx *Context, parentIno uint64, name string, newParentIno uint64, newName string) fuse.Status
	Link(ctx *Context, ino uint64, newParentIno uint64, newName string) fuse.Status

	// Symlinks
	Symlink(ctx *Context, parentIno uint64, name string, target string) fuse.Status
	Readlink(ctx *Context, ino uint64) (target string, code fuse.Status)

	// Extended attributes
	GetXAttr(ctx *Context, ino uint64, attr string) (data []byte, code fuse.Status)
	ListXAttr(ctx *Context, ino uint64) (attrs []string, code fuse.Status)
	SetXAttr(ctx *Context, ino uint64, attr string, data []byte, flags uint32) fuse.Status
	RemoveXAttr(ctx *Context, ino uint64, attr string) fuse.Status

	// File
	Create(ctx *Context, parentIno uint64, name string, flags uint32, mode uint32) (uFh uint32, forceDIO bool, code fuse.Status)
	Open(ctx *Context, ino uint64, flags uint32) (uFh uint32, keepCache, forceDIO bool, code fuse.Status)

	Read(ctx *Context, ino uint64, uFh uint32, dest []byte, off uint64) (result fuse.ReadResult, code fuse.Status)
	Write(ctx *Context, ino uint64, uFh uint32, data []byte, off uint64) (written uint32, code fuse.Status)
	Fallocate(ctx *Context, ino uint64, uFh uint32, off uint64, size uint64, mode uint32) fuse.Status
	Fsync(ctx *Context, ino uint64, uFh uint32, flags uint32) fuse.Status
	Flush(ctx *Context, ino uint64, uFh uint32) fuse.Status
	Release(ctx *Context, ino uint64, uFh uint32)

	GetLk(ctx *Context, ino uint64, uFh uint32, owner uint64, lk *fuse.FileLock, flags uint32, out *fuse.FileLock) fuse.Status
	SetLk(ctx *Context, ino uint64, uFh uint32, owner uint64, lk *fuse.FileLock, flags uint32) fuse.Status
	SetLkw(ctx *Context, ino uint64, uFh uint32, owner uint64, lk *fuse.FileLock, flags uint32) fuse.Status

	// uFh may be 0.
	Chmod(ctx *Context, ino uint64, uFh uint32, mode uint32) fuse.Status
	Chown(ctx *Context, ino uint64, uFh uint32, uid uint32, gid uint32) fuse.Status
	Truncate(ctx *Context, ino uint64, uFh uint32, size uint64) fuse.Status
	Utimens(ctx *Context, ino uint64, uFh uint32, atime *time.Time, mtime *time.Time) fuse.Status

	// Directory
	Lsdir(ctx *Context, ino uint64) (stream []fuse.DirEntry, code fuse.Status)

	StatFs(ctx *Context, ino uint64, out *fuse.StatfsOut) fuse.Status
}

// Options sets options for the entire filesystem
type Options struct {
	// If set to nonnil, this defines the overall entry timeout
	// for the file system. See fuse.EntryOut for more information.
	EntryTimeout *time.Duration

	// If set to nonnil, this defines the overall attribute
	// timeout for the file system. See fuse.EntryOut for more
	// information.
	AttrTimeout *time.Duration

	// If set to nonnil, this defines the overall entry timeout
	// for failed lookups (fuse.ENOENT). See fuse.EntryOut for
	// more information.
	NegativeTimeout *time.Duration

	// NullPermissions if set, leaves null file permissions
	// alone. Otherwise, they are set to 755 (dirs) or 644 (other
	// files.), which is necessary for doing a chdir into the FUSE
	// directories.
	NullPermissions bool

	// If nonzero, replace default (zero) UID with the given UID
	UID uint32

	// If nonzero, replace default (zero) GID with the given GID
	GID uint32

	// Logger is a sink for diagnostic messages. Diagnostic
	// messages are printed under conditions where we cannot
	// return error, but want to signal something seems off
	// anyway. If unset, no messages are printed.
	Logger *log.Logger
}
