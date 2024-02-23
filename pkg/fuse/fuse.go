package fuse

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

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

// TODO: Implement the Meta interface
 
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
