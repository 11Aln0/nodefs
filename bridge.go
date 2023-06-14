package nodefs

import (
	"github.com/hanwen/go-fuse/v2/fuse"
	"sync"
	"syscall"
	"time"
)

type rawBridge struct {
	fs      FileSystem
	options Options

	mu sync.Mutex

	files     []*fileEntry
	freeFiles []uint32
}

// NewNodeFS creates a path based filesystem.
func NewNodeFS(fs FileSystem, options *Options) fuse.RawFileSystem {
	if options == nil {
		oneSec := time.Second
		options = &Options{
			EntryTimeout: &oneSec,
			AttrTimeout:  &oneSec,
		}
	}

	b := &rawBridge{
		fs:      fs,
		options: *options,
		//root:    newInode(1, true),
	}

	return b
}

func (b *rawBridge) logf(format string, args ...interface{}) {
	if b.options.Logger != nil {
		b.options.Logger.Printf(format, args...)
	}
}

func (b *rawBridge) setEntryOut(out *fuse.EntryOut) {
	b.setAttrInner(&out.Attr)
}

func (b *rawBridge) setEntryOutTimeout(out *fuse.EntryOut) {
	if b.options.AttrTimeout != nil {
		out.SetAttrTimeout(*b.options.AttrTimeout)
	}
	if b.options.EntryTimeout != nil {
		out.SetEntryTimeout(*b.options.EntryTimeout)
	}
}

func (b *rawBridge) setAttr(out *fuse.AttrOut) {
	b.setAttrInner(&out.Attr)
}

func (b *rawBridge) setAttrTimeout(out *fuse.AttrOut) {
	if b.options.AttrTimeout != nil {
		out.SetTimeout(*b.options.AttrTimeout)
	}
}

func (b *rawBridge) setAttrInner(out *fuse.Attr) {
	if !b.options.NullPermissions && out.Mode&07777 == 0 {
		out.Mode |= 0644
		if out.Mode&syscall.S_IFDIR != 0 {
			out.Mode |= 0111
		}
	}
	if b.options.UID != 0 && out.Uid == 0 {
		out.Uid = b.options.UID
	}
	if b.options.GID != 0 && out.Gid == 0 {
		out.Gid = b.options.GID
	}
	setBlocks(out)
}

func (b *rawBridge) Init(s *fuse.Server) {}

func (b *rawBridge) String() string {
	return "pathfs"
}

func (b *rawBridge) SetDebug(debug bool) {}

func (b *rawBridge) Access(cancel <-chan struct{}, input *fuse.AccessIn) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	return b.fs.Access(ctx, input.NodeId, input.Mask)
}

func (b *rawBridge) Lookup(cancel <-chan struct{}, header *fuse.InHeader, name string, out *fuse.EntryOut) fuse.Status {
	ctx := newContext(cancel, header.Caller)
	defer releaseContext(ctx)

	code := b.lookup(ctx, header.NodeId, name, out)
	if !code.Ok() && b.options.NegativeTimeout != nil {
		out.SetEntryTimeout(*b.options.NegativeTimeout)
	}

	return code
}

func (b *rawBridge) lookup(ctx *Context, ino uint64, name string, out *fuse.EntryOut) fuse.Status {
	code := b.fs.Lookup(ctx, ino, name, out)
	if !code.Ok() {
		return code
	}

	b.setEntryOut(out)
	b.setEntryOutTimeout(out)
	return fuse.OK
}

func (b *rawBridge) Forget(nodeid, nlookup uint64) {
	b.fs.Forget(nodeid, nlookup)
}

func (b *rawBridge) GetAttr(cancel <-chan struct{}, input *fuse.GetAttrIn, out *fuse.AttrOut) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	f := b.file(uint32(input.Fh()), ctx)

	return b.getAttr(ctx, input.NodeId, f.uFh, out)
}

func (b *rawBridge) getAttr(ctx *Context, ino uint64, uFh uint32, out *fuse.AttrOut) fuse.Status {
	code := b.fs.GetAttr(ctx, ino, uFh, &out.Attr)
	if !code.Ok() {
		return code
	}

	b.setAttr(out)
	b.setAttrTimeout(out)
	return fuse.OK
}

func (b *rawBridge) SetAttr(cancel <-chan struct{}, input *fuse.SetAttrIn, out *fuse.AttrOut) (code fuse.Status) {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	fh, _ := input.GetFh()
	f := b.file(uint32(fh), ctx)
	ino := input.NodeId

	if perms, ok := input.GetMode(); ok {
		code = b.fs.Chmod(ctx, ino, f.uFh, perms)
	}

	uid, uok := input.GetUID()
	gid, gok := input.GetGID()
	if code.Ok() && (uok || gok) {
		code = b.fs.Chown(ctx, ino, f.uFh, uid, gid)
	}

	if sz, ok := input.GetSize(); code.Ok() && ok {
		code = b.fs.Truncate(ctx, ino, f.uFh, sz)
	}

	atime, aok := input.GetATime()
	mtime, mok := input.GetMTime()
	if code.Ok() && (aok || mok) {
		var a, m *time.Time
		if aok {
			a = &atime
		}
		if mok {
			m = &mtime
		}
		code = b.fs.Utimens(ctx, ino, f.uFh, a, m)
	}

	if !code.Ok() {
		return code
	}

	return b.getAttr(ctx, ino, f.uFh, out)
}

func (b *rawBridge) Mknod(cancel <-chan struct{}, input *fuse.MknodIn, name string, out *fuse.EntryOut) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	code := b.fs.Mknod(ctx, input.NodeId, name, input.Mode, input.Rdev)
	if !code.Ok() {
		return code
	}
	return b.lookup(ctx, input.NodeId, name, out)
}

func (b *rawBridge) Mkdir(cancel <-chan struct{}, input *fuse.MkdirIn, name string, out *fuse.EntryOut) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	code := b.fs.Mkdir(ctx, input.NodeId, name, input.Mode)
	if !code.Ok() {
		return code
	}
	return b.lookup(ctx, input.NodeId, name, out)
}

func (b *rawBridge) Unlink(cancel <-chan struct{}, header *fuse.InHeader, name string) fuse.Status {
	ctx := newContext(cancel, header.Caller)
	defer releaseContext(ctx)

	return b.fs.Unlink(ctx, header.NodeId, name)
}

func (b *rawBridge) Rmdir(cancel <-chan struct{}, header *fuse.InHeader, name string) fuse.Status {
	ctx := newContext(cancel, header.Caller)
	defer releaseContext(ctx)

	return b.fs.Rmdir(ctx, header.NodeId, name)
}

func (b *rawBridge) Rename(cancel <-chan struct{}, input *fuse.RenameIn, name string, newName string) fuse.Status {
	if input.Flags != 0 {
		return fuse.ENOSYS
	}

	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	return b.fs.Rename(ctx, input.NodeId, name, input.Newdir, newName)
}

func (b *rawBridge) Link(cancel <-chan struct{}, input *fuse.LinkIn, name string, out *fuse.EntryOut) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	code := b.fs.Link(ctx, input.Oldnodeid, input.NodeId, name)
	if !code.Ok() {
		return code
	}

	return b.lookup(ctx, input.NodeId, name, out)
}

func (b *rawBridge) Symlink(cancel <-chan struct{}, header *fuse.InHeader, target string, name string, out *fuse.EntryOut) fuse.Status {
	ctx := newContext(cancel, header.Caller)
	defer releaseContext(ctx)

	code := b.fs.Symlink(ctx, header.NodeId, name, target)
	if !code.Ok() {
		return code
	}

	return b.lookup(ctx, header.NodeId, name, out)
}

func (b *rawBridge) Readlink(cancel <-chan struct{}, header *fuse.InHeader) ([]byte, fuse.Status) {
	ctx := newContext(cancel, header.Caller)
	defer releaseContext(ctx)

	target, code := b.fs.Readlink(ctx, header.NodeId)
	return []byte(target), code
}

func (b *rawBridge) GetXAttr(cancel <-chan struct{}, header *fuse.InHeader, attr string, dest []byte) (uint32, fuse.Status) {
	ctx := newContext(cancel, header.Caller)
	defer releaseContext(ctx)

	data, code := b.fs.GetXAttr(ctx, header.NodeId, attr)
	if !code.Ok() {
		return 0, code
	}

	sz := len(data)
	if sz > len(dest) {
		return uint32(sz), fuse.ERANGE
	}

	copy(dest, data)
	return uint32(sz), fuse.OK
}

func (b *rawBridge) ListXAttr(cancel <-chan struct{}, header *fuse.InHeader, dest []byte) (uint32, fuse.Status) {
	ctx := newContext(cancel, header.Caller)
	defer releaseContext(ctx)

	attrs, code := b.fs.ListXAttr(ctx, header.NodeId)
	if !code.Ok() {
		return 0, code
	}

	sz := 0
	for _, v := range attrs {
		sz += len(v) + 1
	}
	if sz > len(dest) {
		return uint32(sz), fuse.ERANGE
	}

	dest = dest[:0]
	for _, v := range attrs {
		dest = append(dest, v...)
		dest = append(dest, 0)
	}
	return uint32(sz), fuse.OK
}

func (b *rawBridge) SetXAttr(cancel <-chan struct{}, input *fuse.SetXAttrIn, attr string, data []byte) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	return b.fs.SetXAttr(ctx, input.NodeId, attr, data, input.Flags)
}

func (b *rawBridge) RemoveXAttr(cancel <-chan struct{}, header *fuse.InHeader, attr string) fuse.Status {
	ctx := newContext(cancel, header.Caller)
	defer releaseContext(ctx)

	return b.fs.RemoveXAttr(ctx, header.NodeId, attr)
}

func (b *rawBridge) Create(cancel <-chan struct{}, input *fuse.CreateIn, name string, out *fuse.CreateOut) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	uFh, forceDIO, code := b.fs.Create(ctx, input.NodeId, name, input.Flags, input.Mode)
	if !code.Ok() {
		return code
	}
	code = b.lookup(ctx, input.NodeId, name, &out.EntryOut)
	if !code.Ok() {
		return code
	}
	if forceDIO {
		out.OpenFlags |= fuse.FOPEN_DIRECT_IO
	}
	out.Fh = uint64(b.registerFile(input.Caller.Owner, input.NodeId, uFh, nil))
	return fuse.OK
}

func (b *rawBridge) Open(cancel <-chan struct{}, input *fuse.OpenIn, out *fuse.OpenOut) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	uFh, keepCache, forceDIO, code := b.fs.Open(ctx, input.NodeId, input.Flags)
	if !code.Ok() {
		return code
	}

	out.Fh = uint64(b.registerFile(input.Caller.Owner, input.NodeId, uFh, nil))
	if forceDIO {
		out.OpenFlags |= fuse.FOPEN_DIRECT_IO
	} else if keepCache {
		out.OpenFlags |= fuse.FOPEN_KEEP_CACHE
	}
	return fuse.OK
}

func (b *rawBridge) Read(cancel <-chan struct{}, input *fuse.ReadIn, dest []byte) (fuse.ReadResult, fuse.Status) {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	f := b.file(uint32(input.Fh), ctx)

	return b.fs.Read(ctx, input.NodeId, f.uFh, dest, input.Offset)
}

func (b *rawBridge) Write(cancel <-chan struct{}, input *fuse.WriteIn, data []byte) (written uint32, status fuse.Status) {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	f := b.file(uint32(input.Fh), ctx)

	return b.fs.Write(ctx, input.NodeId, f.uFh, data, input.Offset)
}

func (b *rawBridge) Fallocate(cancel <-chan struct{}, input *fuse.FallocateIn) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	f := b.file(uint32(input.Fh), ctx)

	return b.fs.Fallocate(ctx, input.NodeId, f.uFh, input.Offset, input.Length, input.Mode)
}

func (b *rawBridge) Fsync(cancel <-chan struct{}, input *fuse.FsyncIn) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	f := b.file(uint32(input.Fh), ctx)

	return b.fs.Fsync(ctx, input.NodeId, f.uFh, input.FsyncFlags)
}

func (b *rawBridge) Flush(cancel <-chan struct{}, input *fuse.FlushIn) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	f := b.file(uint32(input.Fh), ctx)

	return b.fs.Flush(ctx, input.NodeId, f.uFh)
}

func (b *rawBridge) Release(cancel <-chan struct{}, input *fuse.ReleaseIn) {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	f := b.file(uint32(input.Fh), ctx)

	b.fs.Release(ctx, input.NodeId, f.uFh)

	b.unregisterFile(uint32(input.Fh))
}

func (b *rawBridge) GetLk(cancel <-chan struct{}, input *fuse.LkIn, out *fuse.LkOut) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	f := b.file(uint32(input.Fh), ctx)

	return b.fs.GetLk(ctx, input.NodeId, f.uFh, input.Owner, &input.Lk, input.LkFlags, &out.Lk)
}

func (b *rawBridge) SetLk(cancel <-chan struct{}, input *fuse.LkIn) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	f := b.file(uint32(input.Fh), ctx)

	return b.fs.SetLk(ctx, input.NodeId, f.uFh, input.Owner, &input.Lk, input.LkFlags)
}

func (b *rawBridge) SetLkw(cancel <-chan struct{}, input *fuse.LkIn) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	f := b.file(uint32(input.Fh), ctx)

	return b.fs.SetLkw(ctx, input.NodeId, f.uFh, input.Owner, &input.Lk, input.LkFlags)
}

func (b *rawBridge) OpenDir(cancel <-chan struct{}, input *fuse.OpenIn, out *fuse.OpenOut) fuse.Status {
	out.Fh = uint64(b.registerFile(input.Caller.Owner, input.NodeId, 0, nil))
	return fuse.OK
}

func (b *rawBridge) ReadDir(cancel <-chan struct{}, input *fuse.ReadIn, out *fuse.DirEntryList) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	d := b.file(uint32(input.Fh), ctx)

	d.mu.Lock()
	defer d.mu.Unlock()

	// rewinddir() should be as if reopening directory.
	if d.stream == nil || input.Offset == 0 {
		stream, code := b.fs.Lsdir(ctx, input.NodeId)
		if !code.Ok() {
			return code
		}
		d.stream = append(stream,
			fuse.DirEntry{Mode: fuse.S_IFDIR, Name: "."},
			fuse.DirEntry{Mode: fuse.S_IFDIR, Name: ".."})
	}

	if input.Offset > uint64(len(d.stream)) {
		// See https://github.com/hanwen/go-fuse/issues/297
		// This can happen for FUSE exported over NFS.  This
		// seems incorrect, (maybe the kernel is using offsets
		// from other opendir/readdir calls), it is harmless to reinforce that
		// we have reached EOF.
		return fuse.OK
	}

	for _, e := range d.stream[input.Offset:] {
		if e.Name == "" {
			b.logf("warning: got empty directory entry, mode %o.", e.Mode)
			continue
		}

		ok := out.AddDirEntry(e)
		if !ok {
			break
		}
	}
	return fuse.OK
}

func (b *rawBridge) ReadDirPlus(cancel <-chan struct{}, input *fuse.ReadIn, out *fuse.DirEntryList) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	d := b.file(uint32(input.Fh), ctx)
	ino := input.NodeId

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.stream == nil || input.Offset == 0 {
		stream, code := b.fs.Lsdir(ctx, ino)
		if !code.Ok() {
			return code
		}
		d.stream = append(stream,
			fuse.DirEntry{Mode: fuse.S_IFDIR, Name: "."},
			fuse.DirEntry{Mode: fuse.S_IFDIR, Name: ".."})
	}

	if input.Offset > uint64(len(d.stream)) {
		return fuse.OK
	}

	for _, e := range d.stream[input.Offset:] {
		if e.Name == "" {
			b.logf("warning: got empty directory entry, mode %o.", e.Mode)
			continue
		}

		// we have to be sure entry will fit if we try to add
		// it, or we'll mess up the lookup counts.
		entryOut := out.AddDirLookupEntry(e)
		if entryOut == nil {
			break
		}
		// No need to fill attributes for . and ..
		if e.Name == "." || e.Name == ".." {
			continue
		}

		b.lookup(ctx, ino, e.Name, entryOut)
	}
	return fuse.OK
}

func (b *rawBridge) ReleaseDir(input *fuse.ReleaseIn) {
	b.unregisterFile(uint32(input.Fh))
}

func (b *rawBridge) FsyncDir(cancel <-chan struct{}, input *fuse.FsyncIn) fuse.Status {
	return fuse.ENOSYS
}

func (b *rawBridge) Lseek(cancel <-chan struct{}, input *fuse.LseekIn, out *fuse.LseekOut) fuse.Status {
	return fuse.ENOSYS
}

func (b *rawBridge) CopyFileRange(cancel <-chan struct{}, input *fuse.CopyFileRangeIn) (written uint32, code fuse.Status) {
	return 0, fuse.ENOSYS
}

func (b *rawBridge) StatFs(cancel <-chan struct{}, input *fuse.InHeader, out *fuse.StatfsOut) fuse.Status {
	ctx := newContext(cancel, input.Caller)
	defer releaseContext(ctx)

	return b.fs.StatFs(ctx, input.NodeId, out)
}
