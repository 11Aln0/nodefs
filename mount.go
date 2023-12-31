package nodefs

import (
	"github.com/hanwen/go-fuse/v2/fuse"
	"time"
)

// Mount mounts the given PathFS on the directory, and starts serving
// requests. This is a convenience wrapper around NewNodeFS and
// fuse.NewServer.  If nil is given as options, default settings are
// applied, which are 1 second entry and attribute timeout.
func Mount(dir string, fs FileSystem, options *Options, mntOptions *fuse.MountOptions) (*fuse.Server, error) {
	if options == nil {
		oneSec := time.Second
		options = &Options{
			EntryTimeout: &oneSec,
			AttrTimeout:  &oneSec,
		}
	}

	rawFS := NewNodeFS(fs, options)
	server, err := fuse.NewServer(rawFS, dir, mntOptions)
	if err != nil {
		return nil, err
	}

	go server.Serve()
	if err := server.WaitMount(); err != nil {
		// we don't shutdown the serve loop. If the mount does
		// not succeed, the loop won't work and exit.
		return nil, err
	}

	return server, nil
}
