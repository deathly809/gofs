// The concrete package contains concrete implementations
// of the FileSystem interface.

// There are two files used by the concrete file system
// The first file contains all the names of the files and their locations
// The second file contains the file data

package concrete

import (
	"io"

	"github.com/deathly809/gofs"
	"github.com/deathly809/gofs/mmap"
)

// Each file in the FileSystem is represented by a linked
// list structure.  This is the order they are physically
// written
type fileNode struct {
	id   int64
	prev int64  // We _PointerSize
	next int64  // _PointerSize
	data []byte // _BlockSize - 2 * _PointerSize
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

// Each entry contains these values
const (
	_NameLength       = 2
	_NameSize         = 256
	_LengthSize       = 8
	_FirstSize        = 8
	_LastSize         = 8
	_LastModifiedSize = 8
	_CreatedSize      = 8

	_EntrySize = _NameLength + _NameSize + _LengthSize + _FirstSize + _LastSize + _LastModifiedSize + _CreatedSize
)

// Data block values
const (
	// Used to grab a new block
	_NullIndex   = -1
	_BlockSize   = 4096
	_PointerSize = 8
	_DataSize    = _BlockSize - 2*_PointerSize
)

// The actual implementation
type fileSystemImpl struct {
	numFiles         int64                // number of files in the filesystem
	sizeInBytes      int64                // the total number of bytes that the files take up
	indexOfFirstFree int64                // the index of the first free node
	indexOfLastFree  int64                // the index of the last free node
	numberFreeNodes  int64                // The number of nodes on the free list
	safeFiles        map[string]gofs.File // the list of files which are locked
	files            map[string]fileInfo  // the list of files in the file system
	dataFile         mmap.File            // the data file
	nameFile         mmap.File            // the name file
	fsName           string               // name of the file system
	fsDirectory      string               // directory where stored on disk
	rawData          []byte               // underlying data
}

func (fSys *fileSystemImpl) GetSafeWriter(file gofs.File) io.Writer {
	return nil
}

func (fSys *fileSystemImpl) GetSafeReader(file gofs.File) io.Reader {
	return nil
}

func (fSys *fileSystemImpl) Shutdown() {

}

func (fSys *fileSystemImpl) Lock(file gofs.File) {

}

func (fSys *fileSystemImpl) Unlock(file gofs.File) {

}

func (fSys *fileSystemImpl) GetWriter() io.Writer {
	return nil
}

func (fSys *fileSystemImpl) Open(filename string) gofs.File {
	return nil
}

func (fSys *fileSystemImpl) Exists(filename string) bool {
	return false
}

func (fSys *fileSystemImpl) Delete(filename string) {
	info, exists := fSys.files[filename]
	// Remove from list of files,
	if exists {
		//fSys.Lock(filename)
		delete(fSys.files, filename)
		//fSys.Unlock(filename)

		// add the file to the free list
		// fSys.Lock("freelist")
		block := fSys.getBlock(info.last)
		block.next = fSys.indexOfFirstFree
		fSys.indexOfFirstFree = info.first
		// fSys.UnLock("freelist")

	}
}
