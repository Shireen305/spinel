// Ignore existing errors, will have to implement inode handlers later

package vfs

import (
	"syscall"

	"github.com/souvikdeyrit/spinel/pkg/chunk"
	"github.com/souvikdeyrit/spinel/pkg/meta"
	"github.com/souvikdeyrit/spinel/pkg/utils"
)

type Ino = meta.Ino
type Attr = meta.Attr
type Context = LogContext

const (
	rootID      = 1
	maxName     = 255
	maxSymlink  = 4096
	maxFileSize = meta.ChunkSize << 31

	modeRead  = 1
	modeWrite = 2
)

type StorageConfig struct {
	Name       string
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Key        string
	KeyPath    string
	Passphrase string
}

type Config struct {
	Meta       *meta.Config
	Format     *meta.Format
	Primary    *StorageConfig
	Chunk      *chunk.Config
	Version    string
	Mountpoint string
	Prefix     string
}

var (
	m      meta.Meta
	reader DataReader // To be implemented later
	writer DataWriter // to be implemented later
)

func Lookup(ctx Context, parent Ino, name string) (entry *meta.Entry, err syscall.Errno) {
		
}

func GetAttr(ctx Context, ino Ino, opened uint8) (entry *meta.Entry, err syscall.Errno) {
	
}

func Mknod(ctx Context, parent Ino, name string, mode uint16, cumask uint16, rdev uint32) (entry *meta.Entry, err syscall.Errno) {
	
}

func Unlink(ctx Context, parent Ino, name string) (err syscall.Errno) {
	
}

func Mkdir(ctx Context, parent Ino, name string, mode uint16, cumask uint16) (entry *meta.Entry, err syscall.Errno) {

}

func Rmdir(ctx Context, parent Ino, name string) (err syscall.Errno) {
	
}

func Symlink(ctx Context, path string, parent Ino, name string) (entry *meta.Entry, err syscall.Errno) {
	
}

func Readlink(ctx Context, ino Ino) (path []byte, err syscall.Errno) {
	
}

func Rename(ctx Context, parent Ino, name string, newparent Ino, newname string) (err syscall.Errno) {
	
}

func Link(ctx Context, ino Ino, newparent Ino, newname string) (entry *meta.Entry, err syscall.Errno) {
	
}

func UpdateEntry(e *meta.Entry) {
	
}

func Readdir(ctx Context, ino Ino, size uint32, off int, fh uint64, plus bool) (entries []*meta.Entry, err syscall.Errno) {
	
}

func Create(ctx Context, parent Ino, name string, mode uint16, cumask uint16, flags uint32) (entry *meta.Entry, fh uint64, err syscall.Errno) {
	
}

func Open(ctx Context, ino Ino, flags uint32) (entry *meta.Entry, fh uint64, err syscall.Errno) {
	
}

func Truncate(ctx Context, ino Ino, size int64, opened uint8, attr *Attr) (err syscall.Errno) {
	
}

func Read(ctx Context, ino Ino, buf []byte, off uint64, fh uint64) (n int, err syscall.Errno) {
	
}

func Write(ctx Context, ino Ino, buf []byte, off, fh uint64) (err syscall.Errno) {
	
}

func Fallocate(ctx Context, ino Ino, mode uint8, off, length int64, fh uint64) (err syscall.Errno) {
	
}

func Fsync(ctx Context, ino Ino, datasync int, fh uint64) (err syscall.Errno) {
	
}

func SetXattr(ctx Context, ino Ino, name string, value []byte, flags int) (err syscall.Errno) {
	
}

func GetXattr(ctx Context, ino Ino, name string, size uint32) (value []byte, err syscall.Errno) {
	
}

func ListXattr(ctx Context, ino Ino, size int) (data []byte, err syscall.Errno) {
	
}

func RemoveXattr(ctx Context, ino Ino, name string) (err syscall.Errno) {
	
}

var logger = utils.GetLogger("spinel")

func Init(conf *Config, m_ meta.Meta, store chunk.ChunkStore) {
	m = m_
	reader = NewDataReader(conf, m, store)
	writer = NewDataWriter(conf, m, store)
	handles = make(map[Ino][]*handle)
}
