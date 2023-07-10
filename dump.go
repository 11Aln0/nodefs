package nodefs

import "github.com/hanwen/go-fuse/v2/fuse"

type DumpRawBridge struct {
	Files     []*DumpFileEntry
	FreeFiles []uint32
}

type DumpFileEntry struct {
	Opener fuse.Owner
	UFh    uint32
	Stream []fuse.DirEntry
}

type Copier interface {
	Dump() (data *DumpRawBridge, err error)
	Restore(data *DumpRawBridge) error
}
