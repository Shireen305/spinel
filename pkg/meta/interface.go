package meta

import (
	"syscall"
)

const (
	DeleteChunk = 1000
)

type MsgCallback func(...interface{}) error

type Attr struct {
	Flags     uint8
	Typ       uint8
	Mode      uint16
	Uid       uint32
	Gid       uint32
	Atime     int64
	Mtime     int64
	Ctime     int64
	Atimensec uint32
	Mtimensec uint32
	Ctimensec uint32
	Nlink     uint32
	Length    uint64
	Rdev      uint32
	Full      bool
}

type Entry struct {
	Inode Ino
	Name  []byte
	Attr  *Attr
}

type Slice struct {
	Chunkid uint64
	Size    uint32
	Off     uint32
	Len     uint32
}

type Meta interface {
	Init(format Format) error
	Load() (*Format, error)

	StatFS(ctx Context, totalspace, availspace, iused, iavail *uint64) syscall.Errno
	Access(ctx Context, inode Ino, modemask uint16) syscall.Errno
	Lookup(ctx Context, parent Ino, name string, inode *Ino, attr *Attr) syscall.Errno
	GetAttr(ctx Context, inode Ino, attr *Attr) syscall.Errno
	SetAttr(ctx Context, inode Ino, set uint16, sggidclearmode uint8, attr *Attr) syscall.Errno
	Truncate(ctx Context, inode Ino, flags uint8, attrlength uint64, attr *Attr) syscall.Errno
	Fallocate(ctx Context, inode Ino, mode uint8, off uint64, size uint64) syscall.Errno
	ReadLink(ctx Context, inode Ino, path *[]byte) syscall.Errno
	Symlink(ctx Context, parent Ino, name string, path string, inode *Ino, attr *Attr) syscall.Errno
	Mknod(ctx Context, parent Ino, name string, _type uint8, mode uint16, cumask uint16, rdev uint32, inode *Ino, attr *Attr) syscall.Errno
	Mkdir(ctx Context, parent Ino, name string, mode uint16, cumask uint16, copysgid uint8, inode *Ino, attr *Attr) syscall.Errno
	Unlink(ctx Context, parent Ino, name string) syscall.Errno
	Rmdir(ctx Context, parent Ino, name string) syscall.Errno
	Rename(ctx Context, parentSrc Ino, nameSrc string, parentDst Ino, nameDst string, inode *Ino, attr *Attr) syscall.Errno
	Link(ctx Context, inodeSrc, parent Ino, name string, attr *Attr) syscall.Errno
	Readdir(ctx Context, inode Ino, wantattr uint8, entries *[]*Entry) syscall.Errno
	Create(ctx Context, parent Ino, name string, mode uint16, cumask uint16, inode *Ino, attr *Attr) syscall.Errno
	Open(ctx Context, inode Ino, flags uint8, attr *Attr) syscall.Errno
	Close(ctx Context, inode Ino) syscall.Errno
	Read(inode Ino, indx uint32, chunks *[]Slice) syscall.Errno
	NewChunk(ctx Context, inode Ino, indx uint32, offset uint32, chunkid *uint64) syscall.Errno
	Write(ctx Context, inode Ino, indx uint32, off uint32, slice Slice) syscall.Errno

	GetXattr(ctx Context, inode Ino, name string, vbuff *[]byte) syscall.Errno
	ListXattr(ctx Context, inode Ino, dbuff *[]byte) syscall.Errno
	SetXattr(ctx Context, inode Ino, name string, value []byte) syscall.Errno
	RemoveXattr(ctx Context, inode Ino, name string) syscall.Errno
	Flock(ctx Context, inode Ino, owner uint64, ltype uint32, block bool) syscall.Errno
	Getlk(ctx Context, inode Ino, owner uint64, ltype *uint32, start, end *uint64, pid *uint32) syscall.Errno
	Setlk(ctx Context, inode Ino, owner uint64, block bool, ltype uint32, start, end uint64, pid uint32) syscall.Errno

	OnMsg(mtype uint32, cb MsgCallback)
}
