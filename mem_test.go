package nodefs

import (
	"github.com/hanwen/go-fuse/v2/fuse"
	"strconv"
	"syscall"
	"testing"
	"time"
)

// 每个文件都有childCnt个子文件，最深到level层，子文件的name为"dir{子文件序号}", ino为父文件ino * childCnt + 子文件序号
type testFs struct {
	defaultFileSystem
	level    int
	childCnt int
	maxIno   uint64
	fileAttr fuse.Attr
}

func newTestFs(level int, childCnt int) FileSystem {
	now := time.Now()
	fs := &testFs{
		level:    level,
		childCnt: childCnt,
		maxIno:   1,
	}
	// 计算最大Ino
	for i := 1; i < level; i++ {
		fs.maxIno = fs.maxIno*uint64(fs.childCnt) + uint64(fs.childCnt-1)
	}
	// 填充文件Attr
	fs.fileAttr.SetTimes(&now, &now, &now)
	fs.fileAttr.Mode = fuse.S_IFDIR | 0777
	fs.fileAttr.Size = 4
	return fs
}

func (fs *testFs) GetAttr(ctx *Context, ino uint64, uFh uint32, out *fuse.Attr) fuse.Status {
	if ino > fs.maxIno {
		return fuse.ENOENT
	}
	*out = fs.fileAttr
	out.Ino = ino
	return fuse.OK
}

func (fs *testFs) Lookup(ctx *Context, parentIno uint64, name string, out *fuse.Attr) fuse.Status {
	if len(name) < 4 || parentIno*uint64(fs.childCnt) > fs.maxIno {
		return fuse.ENOENT
	}
	if num, err := strconv.Atoi(name[3:]); name[:3] == "dir" && err == nil && num < fs.childCnt {
		*out = fs.fileAttr
		out.Ino = parentIno*uint64(fs.childCnt) + uint64(num)
		return fuse.OK
	} else {
		return fuse.ENOENT
	}
}

func TestMem(t *testing.T) {
	maxLevel := 8
	childCnt := 9 // 5380840个inode
	mountPoint := "/tmp/nodefs"
	fs := newTestFs(maxLevel, childCnt)
	server, err := Mount(mountPoint, fs, nil, nil)
	if err != nil {
		panic(err)
	}
	defer server.Unmount()

	var traversal func(int, string)
	stat := syscall.Stat_t{}
	traversal = func(level int, path string) {
		if level > maxLevel {
			return
		}
		err = syscall.Stat(path, &stat)
		if err != nil {
			panic(err)
		}
		for i := 0; i < childCnt; i++ {
			traversal(level+1, path+"/dir"+string('0'+byte(i)))
		}
	}
	traversal(1, mountPoint)
	return
}
