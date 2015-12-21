// The concrete package contains concrete implementations
// of the FileSystem interface.

package concrete

import (
	"bytes"
	"dfs/fs"
	"dfs/fs/mmap"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// Each file in the FileSystem is represented by a linked
// list structure
type fileNode struct {
	data []byte
	next *fileNode
	prev *fileNode
}

// This is the header for each file in the filesystem
type fileInfo struct {
	pos          int
	size         int
	first        *fileNode
	created      time.Time
	lastModified time.Time
}

type file struct {
	fs   *fileSystemImpl
	pos  int
	head *fileNode
	curr *fileNode
	end  *fileNode
}

func (f *file) Close() error {
	return nil
}

func (f *file) Write(data []byte) (n int, err error) {
	return 0, nil
}

func (f *file) Read(data []byte) (n int, err error) {
	return 0, nil
}

func (f *file) Seek(pos int, off fs.FileOffset) {
	switch off {
	case fs.Beginning:
		f.curr = f.head
	case fs.End:
		f.curr = f.head
	}

	if pos < 0 { // Back
		if f.curr.prev == nil {
			t := f.fs.getBlock()
			t.next = f.curr
			f.curr.prev = t
			t.next = f.curr
			f.curr = t
		}
	} else { // Forward

	}
}

func (f *file) IsNew() bool {
	return false
}

func (f *file) Name() string {
	return ""
}

func (f *file) Size() int {
	return 0
}

const (
	// Major version of the filesystem
	Major = int32(0)
	// Minor version of the filesystem
	Minor = int32(1)
	// Patch version of the filesystem
	Patch = int32(0)
)

// 	The header layout contains a signature, version, number of files, filesystem size,
//	and first block of free list.
// signature 	= 8 bytes
// version   	= 12 bytes
// number files = 16 bytes
// size			= 16 bytes
// first free	= 16 bytes
var _Signature = []byte{0xD, 0xE, 0xA, 0xD, 0xB, 0xE, 0xE, 0xF}

const (
	_SignatureSize  = 8
	_VersionBytes   = 12
	_FileCountBytes = 12
	_SizeBytes      = 12
	_FirstFreeBytes = 12
	_MajorVersion   = 2

	_HeaderSize = _SignatureSize + _VersionBytes + _FileCountBytes + _SizeBytes + _FirstFreeBytes
)

// The actual implementation
type fileSystemImpl struct {
	numFiles  int64
	firstFree fileNode
	safeFiles map[string]fs.File
	files     map[string]fileInfo
	mFile     fs.File
}

func (fSys *fileSystemImpl) readHeader() error {
	header := make([]byte, _HeaderSize)

	fSys.mFile.Seek(0, fs.Beginning)
	fSys.mFile.Read(header)

	buffer := bytes.NewReader(header)

	signature := make([]byte, _SignatureSize)
	buffer.Read(signature)

	var major, minor, patch int32
	binary.Read(buffer, binary.BigEndian, &major)
	binary.Read(buffer, binary.BigEndian, &minor)
	binary.Read(buffer, binary.BigEndian, &patch)

	if major != Major {
		msg := fmt.Sprintf("Trying to load an incompatible filesystem version: %d.%d.%d", major, minor, patch)
		return errors.New(msg)
	}

	var numFiles, size, firstFree int32
	binary.Read(buffer, binary.BigEndian, &numFiles)
	binary.Read(buffer, binary.BigEndian, &size)
	binary.Read(buffer, binary.BigEndian, &firstFree)

	return nil
}

func (fSys *fileSystemImpl) GetSafeWriter(file fs.File) io.Writer {
	return nil
}

func (fSys *fileSystemImpl) GetSafeReader(file fs.File) io.Reader {
	return nil
}

func (fSys *fileSystemImpl) Shutdown() {

}

func (fSys *fileSystemImpl) Lock(file fs.File) {

}

func (fSys *fileSystemImpl) Unlock(file fs.File) {

}

func (fSys *fileSystemImpl) GetWriter() io.Writer {
	return nil
}

func (fSys *fileSystemImpl) Open(filename string) fs.File {
	return nil
}

func (fSys *fileSystemImpl) Exists(filename string) bool {
	return false
}

func (fSys *fileSystemImpl) Delete(filename string) {
	info, exists := fSys.files[filename]
	// Remove from list of files,
	if exists {
		fSys.Lock(filename)
		info.pos = 0
		fSys.Unlock(filename)
	}
}

// Initializes the filesystem after the MMAPFile has been
// opened
func (fSys *fileSystemImpl) init() error {
	//	Read filesystem header
	err := fSys.readHeader()
	return err
}

func (fSys *fileSystemImpl) getBlock() *fileNode {
	return nil
}

// Open creates the default filesystem
func Open(filename string) (fs.FileSystem, error) {
	var err error

	result := new(fileSystemImpl)

	result.mFile, err = mmap.NewFile(filename)

	if err != nil {
		return nil, errors.New("Could not open filesystem: " + err.Error())
	}

	return result, nil
}
