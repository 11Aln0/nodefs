package nodefs

import (
	"github.com/hanwen/go-fuse/v2/fuse"
	"testing"
)

func TestRegister(t *testing.T) {
	b := &rawBridge{
		fs:    DefaultFileSystem(),
		files: []*fileEntry{{}},
	}

	fh := b.registerFile(fuse.Owner{}, 1, nil)

	b.unregisterFile(fh)
	if len(b.freeFiles) != 1 {
		t.Errorf("want freeFiles count: %d, have: %d", 1, len(b.freeFiles))
	}

	fh = b.registerFile(fuse.Owner{}, b2, nil)
	if len(b.freeFiles) != 0 {
		t.Errorf("want freeFiles count: %d, have: %d", 0, len(b.freeFiles))
	}

}
