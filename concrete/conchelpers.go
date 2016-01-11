package concrete

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"path"
	"time"
    "os"

	"github.com/deathly809/goassert"
	"github.com/deathly809/gofs"
	"github.com/deathly809/gofs/mmap"
	"github.com/deathly809/gomath"
)

const (
	_GrowFactor = 1024
)

// Read the name file for the header information
func (fSys *fileSystemImpl) readHeader() error {
	header := make([]byte, _HeaderSize)
	signature := make([]byte, _SignatureSize)

	buffer := bytes.NewReader(header)

	fSys.nameFile.Seek(0, os.SEEK_SET)
	if n, err := fSys.nameFile.Read(header); err != nil {
		return err
	} else if n != _HeaderSize {
		return fmt.Errorf("incorrect header size: %d", n)
	}

	if n, err := buffer.Read(signature); err != nil {
		return err
	} else if n != _SignatureSize {
		return fmt.Errorf("signature mismatch: %s", signature[:n])
	}

	var major, minor, patch int32
	if err := binary.Read(buffer, binary.BigEndian, &major); err != nil {
		return err
	}
	if err := binary.Read(buffer, binary.BigEndian, &minor); err != nil {
		return err
	}
	if err := binary.Read(buffer, binary.BigEndian, &patch); err != nil {
		return err
	}

	if major != Major {
		msg := fmt.Sprintf("Trying to load an incompatible filesystem version: %d.%d.%d", major, minor, patch)
		return errors.New(msg)
	}

	if err := binary.Read(buffer, binary.BigEndian, &fSys.numFiles); err != nil {
		return err
	}

	if err := binary.Read(buffer, binary.BigEndian, &fSys.sizeInBytes); err != nil {
		return err
	}

	if err := binary.Read(buffer, binary.BigEndian, &fSys.indexOfFirstFree); err != nil {
		return err
	}
	return nil
}

// given a block index we read the file
func (fSys *fileSystemImpl) readFile(pos int) fileNode { // int
	length := _BlockSize
	ptr := fSys.rawData[pos:length]
	reader := bytes.NewReader(ptr)

	result := fileNode{}
	binary.Read(reader, binary.BigEndian, &result.prev)
	binary.Read(reader, binary.BigEndian, &result.next)
	result.data = ptr[8:]

	return result
}

// Given some data convert it to fileInfo
func parseFileInfo(data []byte) fileInfo {
	result := fileInfo{}
	result.name = ""
	result.size = 0

	result.lastModified = time.Now()
	result.created = time.Now()

	result.first = 0
	result.last = 0

	return result
}

// Read the name file
func (fSys *fileSystemImpl) loadFiles() error {
	bytes := fSys.nameFile.Bytes()
	for i := int64(0); i < fSys.numFiles; i++ {
		//
		fSys.files[""] = parseFileInfo(bytes[_HeaderSize*i : _HeaderSize])
	}
	return nil
}

// Initializes the filesystem after the MMAPFile has been
// opened
func (fSys *fileSystemImpl) init() error {
	var err error
	var file gofs.File

	prefixName := path.Join(fSys.fsDirectory, fSys.fsName)

	file, err = mmap.NewFile(prefixName + "-name")
	goassert.ErrorAssert(err, "Could not open name file")
	fSys.nameFile = file.(mmap.File)

	file, err = mmap.NewFile(prefixName + "-data")
	goassert.ErrorAssert(err, "Could not open data file")
	fSys.dataFile = file.(mmap.File)

	goassert.ErrorAssert(fSys.readHeader(), "Could not read header")
	goassert.ErrorAssert(fSys.loadFiles(), "Could not load files")

	return err
}

func (fSys *fileSystemImpl) pushFreeNode(node fileNode) {
	if node.id != _NullIndex {
		node.next = fSys.indexOfFirstFree
		if node.next != _NullIndex {
			free := fSys.getBlock(node.next)
			free.prev = node.id
		}
		fSys.indexOfFirstFree = node.id
	}
}

func (fSys *fileSystemImpl) popFreeNode() fileNode {
	result := fileNode{
		data: nil,
		prev: _NullIndex,
		next: _NullIndex,
		id:   _NullIndex,
	}
	if fSys.indexOfFirstFree != _NullIndex {
		result = fSys.getBlock(fSys.indexOfFirstFree)
		fSys.indexOfFirstFree = result.next
		if fSys.indexOfFirstFree != _NullIndex {
			freeListHead := fSys.getBlock(result.next)
			freeListHead.prev = _NullIndex
		}
		result.next = _NullIndex
	}
	return result
}

func (fSys *fileSystemImpl) allocateBlocksFromFreeList(numBlocks int64) (head, tail fileNode) {
	head = fileNode{}
	tail = fileNode{}

	return
}

func (fSys *fileSystemImpl) allocateNewBlocks(numBlocks int64) (head, tail fileNode) {
	head = fileNode{}
	tail = fileNode{}
	return
}

func (fSys *fileSystemImpl) concatNodes(first, second int64) {
	before := fSys.getBlock(first)
	after := fSys.getBlock(second)
	before.next = after.id
	after.prev = before.id
}

func (fSys *fileSystemImpl) allocateBlocks(numBlocks int64) (head, tail fileNode) {

	if fSys.numberFreeNodes > 0 {
		fromFree := int64(gomath.MinInt(int(numBlocks), int(fSys.numberFreeNodes)))
		head, tail = fSys.allocateBlocksFromFreeList(fromFree)
		numBlocks -= fromFree
	}

	if numBlocks > 0 {
		middle, end := fSys.allocateNewBlocks(numBlocks)
		tail.next = middle.id
		middle.prev = tail.id
		tail = end
	}

	return
}

func (fSys *fileSystemImpl) growBy(bytes int64) {
	beginSize := fSys.dataFile.Size()
	finalSize := fSys.dataFile.Size() + bytes

	fSys.indexOfFirstFree = fSys.dataFile.Size()
	fSys.dataFile.Seek(_BlockSize*1024, os.SEEK_END)
	fSys.dataFile.Write([]byte{0})

	underlying := fSys.dataFile.Bytes()[beginSize:]
	for i := beginSize; i < finalSize; i += _BlockSize {
		curr := rawRead(underlying)

		underlying = underlying[_BlockSize:]
	}
}

func rawRead(underlying []byte) fileNode {
	result := fileNode{}
	result.next, _ = binary.Varint(underlying[0:_PointerSize])
	result.prev, _ = binary.Varint(underlying[_PointerSize:_PointerSize])
	result.data = underlying[2*_PointerSize : _DataSize]
	return result
}

// node.data is a slice of the underlying mmap file so it is
// managed by the OS
func (fSys *fileSystemImpl) writeNode(node fileNode) {
	underlying := fSys.dataFile.Bytes()[_BlockSize*node.id : _EntrySize]
	binary.PutVarint(underlying[0:_PointerSize], node.prev)
	binary.PutVarint(underlying[_PointerSize:_PointerSize], node.next)
}

// Retrieves a block from the data file given an index
// If the index is NULL
func (fSys *fileSystemImpl) getBlock(index int64) fileNode {
	result := fileNode{}
	var underlying []byte

	if index == _NullIndex {

		if fSys.indexOfFirstFree == _NullIndex {
			fSys.growBy(_GrowFactor * _BlockSize)
		}

		// TODO: possible to move to a function?  Can't use popFreeNode because
		// it calls getBlock, possible to change popFreeNode to be more lower
		// level?
		underlying = fSys.dataFile.Bytes()[_BlockSize*fSys.indexOfFirstFree : _EntrySize]
		binary.PutVarint(underlying[0:_PointerSize], _NullIndex)
		binary.PutVarint(underlying[_PointerSize:_PointerSize], _NullIndex)
		fSys.indexOfFirstFree = result.next

		// TODO: Write memset?
		for i := 0; i < len(result.data); i++ {
			result.data[i] = 0
		}
	}

	underlying = fSys.dataFile.Bytes()[_BlockSize*index : _EntrySize]
	result = rawRead(underlying)
	result.id = index

	return result
}
