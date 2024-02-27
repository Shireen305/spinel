package fuse

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/souvikdeyrit/spinel/pkg/meta"
	"github.com/souvikdeyrit/spinel/pkg/utils"
	"github.com/souvikdeyrit/spinel/pkg/vfs"

	"github.com/hanwen/go-fuse/v2/fuse"
)

var logger = utils.GetLogger("spinel")

type JFS struct {
	fuse.RawFileSystem
	cacheMode       int
	attrTimeout     time.Duration
	direntryTimeout time.Duration
	entryTimeout    time.Duration
}

func NewJFS() *JFS {
	return &JFS{
		RawFileSystem: fuse.NewDefaultRawFileSystem(),
	}
}

func (fs *JFS) replyEntry(out *fuse.EntryOut, entry *meta.Entry) fuse.Status {
	out.NodeId = uint64(entry.Inode)
	out.Generation = 1
	out.SetAttrTimeout(fs.attrTimeout)
	if entry.Attr.Typ == meta.TypeDirectory {
		out.SetEntryTimeout(fs.direntryTimeout)
	} else {
		out.SetEntryTimeout(fs.entryTimeout)
	}
	if vfs.IsSpecialNode(entry.Inode) {
		out.SetAttrTimeout(time.Hour)
	}
	attrToStat(entry.Inode, entry.Attr, &out.Attr)
	return 0
}

func (fs *JFS) Lookup(cancel <-chan struct{}, header *fuse.InHeader, name string, out *fuse.EntryOut) (status fuse.Status) {
	ctx := newContext(cancel, header)
	defer releaseContext(ctx)
	entry, err := vfs.Lookup(ctx, Ino(header.NodeId), name)
	if err != 0 {
		return fuse.Status(err)
	}
	return fs.replyEntry(out, entry)
}

func (fs *JFS) GetAttr(cancel <-chan struct{}, in *fuse.GetAttrIn, out *fuse.AttrOut) (code fuse.Status) {
	ctx := newContext(cancel, &in.InHeader)
	defer releaseContext(ctx)
	var opened uint8
	if in.Fh() != 0 {
		opened = 1
	}
	entry, err := vfs.GetAttr(ctx, Ino(in.NodeId), opened)
	if err != 0 {
		return fuse.Status(err)
	}
	attrToStat(entry.Inode, entry.Attr, &out.Attr)
	out.AttrValid = uint64(fs.attrTimeout.Seconds())
	if vfs.IsSpecialNode(Ino(in.NodeId)) {
		out.AttrValid = 3600
	}
	return 0
}

func (fs *JFS) Mknod(cancel <-chan struct{}, in *fuse.MknodIn, name string, out *fuse.EntryOut) (code fuse.Status) {
	ctx := newContext(cancel, &in.InHeader)
	defer releaseContext(ctx)
	entry, err := vfs.Mknod(ctx, Ino(in.NodeId), name, uint16(in.Mode), getUmask(in), in.Rdev)
	if err != 0 {
		return fuse.Status(err)
	}
	return fs.replyEntry(out, entry)
}

func (fs *JFS) Mkdir(cancel <-chan struct{}, in *fuse.MkdirIn, name string, out *fuse.EntryOut) (code fuse.Status) {
	ctx := newContext(cancel, &in.InHeader)
	defer releaseContext(ctx)
	entry, err := vfs.Mkdir(ctx, Ino(in.NodeId), name, uint16(in.Mode), uint16(in.Umask))
	if err != 0 {
		return fuse.Status(err)
	}
	return fs.replyEntry(out, entry)
}

/* TODO: Implement the RawFileSystem interface (https://pkg.go.dev/github.com/hanwen/go-fuse/v2@v2.0.3/fuse#RawFileSystem) 
except String, SetDebug, Forget, Lseek, CopyFileRange, FsyncDir and Init (They have default implementations in place) */

func Main(conf *vfs.Config, options string, attrcacheto_, entrycacheto_, direntrycacheto_ float64) error {
	syscall.Setpriority(syscall.PRIO_PROCESS, os.Getpid(), -19)

	imp := NewJFS()
	imp.attrTimeout = time.Millisecond * time.Duration(attrcacheto_*1000)
	imp.entryTimeout = time.Millisecond * time.Duration(entrycacheto_*1000)
	imp.direntryTimeout = time.Millisecond * time.Duration(direntrycacheto_*1000)

	var opt fuse.MountOptions
	opt.FsName = "Spinel:" + conf.Format.Name
	opt.Name = "spinel"
	opt.SingleThreaded = false
	opt.MaxBackground = 50
	opt.EnableLocks = true
	opt.DisableXAttrs = false
	opt.IgnoreSecurityLabels = true
	opt.MaxWrite = 1 << 20
	opt.MaxReadAhead = 1 << 20
	opt.DirectMount = true
	opt.AllowOther = os.Getuid() == 0
	for _, n := range strings.Split(options, ",") {
		if n == "allow_other" || n == "allow_root" {
			opt.AllowOther = true
		} else if strings.HasPrefix(n, "fsname=") {
			opt.FsName = n[len("fsname="):]
			if runtime.GOOS == "darwin" {
				opt.Options = append(opt.Options, "volname="+n[len("fsname="):])
			}
		} else if n == "nonempty" {
		} else if n == "debug" {
			opt.Debug = true
		} else if strings.TrimSpace(n) != "" {
			opt.Options = append(opt.Options, n)
		}
	}
	opt.Options = append(opt.Options, "default_permissions")
	if runtime.GOOS == "darwin" {
		opt.Options = append(opt.Options, "fssubtype=spinel")
		opt.Options = append(opt.Options, "daemon_timeout=60", "iosize=65536", "novncache")
		imp.cacheMode = 2
	}
	fssrv, err := fuse.NewServer(imp, conf.Mountpoint, &opt)
	if err != nil {
		return fmt.Errorf("fuse: %s", err)
	}

	fssrv.Serve()
	return nil
}
