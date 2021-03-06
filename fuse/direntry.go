package fuse

// all of the code for DirEntryList.

import (
	"fmt"
	"unsafe"
)

var eightPadding [8]byte

const direntSize = int(unsafe.Sizeof(_Dirent{}))

// DirEntry is a type for PathFileSystem and NodeFileSystem to return
// directory contents in.
type DirEntry struct {
	Mode uint32
	Name string
}

func (d DirEntry) String() string {
	return fmt.Sprintf("%o: %q", d.Mode, d.Name)
}

// DirEntryList holds the return value for READDIR and READDIRPLUS
// opcodes.
type DirEntryList struct {
	buf    []byte
	size   int
	offset uint64
}

// NewDirEntryList creates a DirEntryList with the given data buffer
// and offset.
func NewDirEntryList(data []byte, off uint64) *DirEntryList {
	return &DirEntryList{
		buf:    data[:0],
		size:   len(data),
		offset: off,
	}
}

// AddDirEntry tries to add an entry, and reports whether it
// succeeded.
func (l *DirEntryList) AddDirEntry(e DirEntry) (bool, uint64) {
	return l.Add(nil, e.Name, uint64(FUSE_UNKNOWN_INO), e.Mode)
}

// Add adds a direntry to the DirEntryList, returning whether it
// succeeded.
func (l *DirEntryList) Add(prefix []byte, name string, inode uint64, mode uint32) (bool, uint64) {
	padding := (8 - len(name)&7) & 7
	delta := padding + direntSize + len(name) + len(prefix)
	oldLen := len(l.buf)
	newLen := delta + oldLen

	if newLen > l.size {
		return false, l.offset
	}
	l.buf = l.buf[:newLen]
	copy(l.buf[oldLen:], prefix)
	oldLen += len(prefix)
	dirent := (*_Dirent)(unsafe.Pointer(&l.buf[oldLen]))
	dirent.Off = l.offset + 1
	dirent.Ino = inode
	dirent.NameLen = uint32(len(name))
	dirent.Typ = (mode & 0170000) >> 12
	oldLen += direntSize
	copy(l.buf[oldLen:], name)
	oldLen += len(name)

	if padding > 0 {
		copy(l.buf[oldLen:], eightPadding[:padding])
	}

	l.offset = dirent.Off
	return true, l.offset
}

// AddDirLookupEntry is used for ReadDirPlus. It serializes a DirEntry
// and its corresponding lookup. Pass a zero entryOut if the lookup
// data should be ignored.
func (l *DirEntryList) AddDirLookupEntry(e DirEntry, entryOut *EntryOut) (bool, uint64) {
	ino := uint64(FUSE_UNKNOWN_INO)
	if entryOut.Ino > 0 {
		ino = entryOut.Ino
	}
	var lookup []byte
	toSlice(&lookup, unsafe.Pointer(entryOut), unsafe.Sizeof(EntryOut{}))

	return l.Add(lookup, e.Name, ino, e.Mode)
}

func (l *DirEntryList) bytes() []byte {
	return l.buf
}
