package nodefs

import (
	"github.com/hanwen/go-fuse/v2/fuse"
	"golang.org/x/sys/unix"
	"log"
	"sync"
)

type fileEntry struct {
	opener fuse.Owner

	// file
	uFh uint32

	// dir
	mu     sync.Mutex
	stream []fuse.DirEntry
}

func (b *rawBridge) file(fh uint32, ctx *Context) *fileEntry {
	b.mu.Lock()
	defer b.mu.Unlock()
	f := b.files[fh]
	if f == nil {
		log.Panicf("unknown file %d", fh)
	}
	if fh != 0 {
		ctx.Opener = &f.opener
	}
	return f
}
func (b *rawBridge) registerFile(opener fuse.Owner, uFh uint32, stream []fuse.DirEntry) (fh uint32) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.freeFiles) > 0 {
		last := len(b.freeFiles) - 1
		fh = b.freeFiles[last]
		b.freeFiles = b.freeFiles[:last]
	} else {
		fh = uint32(len(b.files))
		b.files = append(b.files, &fileEntry{})
	}

	entry := b.files[fh]
	entry.opener = opener
	entry.uFh = uFh
	entry.stream = stream
	return
}

func (b *rawBridge) unregisterFile(fh uint32) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if fh == 0 {
		return
	}
	unix.UnixRights()
	b.files[fh] = &fileEntry{}
	b.freeFiles = append(b.freeFiles, fh)
	return
}
