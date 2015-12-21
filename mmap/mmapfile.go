package mmap

import (
	"bytes"
	"encoding/binary"
	"errors"
	"jeff/common"
	"jeff/math"
	"log"
	"os"
	"sync"
	"unsafe"

	"dfs/fs"

	"github.com/edsrzf/mmap-go"
)

// File represents a File which is mapped to some place in memory
type File interface {
	fs.File
	// Bytes returns the underlying memory that the file backs
	// If you want to use this you will need to lock the file
	// before
	Bytes() []byte

	// Lock will lock the file from further reading or writing
	// through the File Read/Write interface
	Lock()

	// Unlock will allow the file to be written and read from using
	// the File Read/Write interface
	Unlock()
}

var _Sanity = []byte{0x0, 0x0, 0xd, 0x1, 0xe, 0x5, 0x0, 0xf, 0xd, 0x0, 0x0, 0xd, 0xa, 0xd, 0x5}

const (
	_Version     byte = byte(1)
	_MaxFileSize      = 1000000000
	_HeaderSize       = int(unsafe.Sizeof(_Version)+unsafe.Sizeof(1)) + 16
	_InitialSize      = 4096 + _HeaderSize
)

/* Implementation */

type header struct {
	sanity []byte
	ver    byte
	mSize  int64
}

type mmapFileImpl struct {
	memmap  mmap.MMap // mFile is just []byte
	mapSize int
	newFile bool
	file    *os.File
	lock    *sync.Mutex
	name    string
	pos     int
}

/* Required for interface */

func (mFile *mmapFileImpl) Bytes() []byte {
	return mFile.memmap
}

// Close cleans up all resources, flushes, and closes the
// memory mapped file
func (mFile *mmapFileImpl) Close() error {
	mFile.writeHeader()
	mFile.memmap.Flush()
	err := mFile.memmap.Unmap()

	if err != nil {
		return err
	}

	err = mFile.file.Close()
	if err != nil {
		return err
	}

	return nil
}

func (mFile *mmapFileImpl) Write(data []byte) (int, error) {
	start := mFile.pos + _HeaderSize
	end := start + len(data)

	mFile.lock.Lock()
	if end > mFile.mapSize {
		mFile.grow(end + _HeaderSize)
	}

	to := mFile.memmap[start:]
	if len(to) < len(data) {
		log.Fatal("Not enough space, didn't we grow?")
	}

	length := copy(to, data)
	mFile.pos = end // we have moved

	mFile.lock.Unlock()
	mFile.memmap.Flush()

	return length, nil
}

func (mFile *mmapFileImpl) Lock() {
	mFile.lock.Lock()
}

func (mFile *mmapFileImpl) Unlock() {
	mFile.lock.Unlock()
}

func (mFile *mmapFileImpl) Seek(pos int, from fs.FileOffset) {
	switch from {
	case fs.Beginning:
		mFile.pos = pos
	case fs.Current:
		mFile.pos = pos + pos
	case fs.End:
		mFile.pos = mFile.mapSize - pos
	}
	mFile.pos = math.MaxInt(0, math.MinInt(mFile.pos, mFile.mapSize-1))
}

func (mFile *mmapFileImpl) Size() int {
	return mFile.mapSize
}

func (mFile *mmapFileImpl) Read(data []byte) (int, error) {
	start := _HeaderSize + mFile.pos
	end := math.MinInt(mFile.mapSize, start+len(data))

	length := end - start

	if end > mFile.mapSize {
		return 0, errors.New("Tried to read beyond end of file")
	}

	if length > _MaxFileSize {
		log.Fatal("File too large")
	}

	check := copy(data, mFile.memmap[start:])

	if check != length {
		log.Fatal("Could not read entire length")
	}

	// Moved on
	mFile.pos = end

	return length, nil
}

// IsNew returns true if this file was new when created, otherwise
// returns false
func (mFile *mmapFileImpl) IsNew() bool {
	return mFile.newFile
}

// Name returns the filename
func (mFile *mmapFileImpl) Name() string {
	return mFile.name
}

/* Required to work */

func (mFile *mmapFileImpl) grow(newSize int) {
	// Get the new size
	mFile.mapSize = newSize

	// Flush and unmap
	mFile.memmap.Flush()
	mFile.memmap.Unmap()
	mFile.memmap = nil

	// Grow the file
	mFile.file.Truncate(int64(mFile.mapSize))

	var err error

	mFile.memmap, err = mmap.Map(mFile.file, mmap.RDWR, 0)

	if err != nil {
		log.Fatal("Could not resize file, handle gracefully later: ", err.Error())
	}

	if mFile.memmap == nil {
		log.Fatal("Could not resize file, handle gracefully later(2)")
	}

	if len(mFile.memmap) != mFile.mapSize {
		log.Fatal("Backing mapped array not same size")
	}

}

func (mFile *mmapFileImpl) writeHeader() {

	var buff bytes.Buffer
	err := binary.Write(&buff, binary.BigEndian, _Sanity)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = binary.Write(&buff, binary.BigEndian, _Version)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = binary.Write(&buff, binary.BigEndian, int64(mFile.mapSize))
	if err != nil {
		log.Fatal(err.Error())
	}
	copy(mFile.memmap, buff.Bytes())
	mFile.memmap.Flush()
}

func (mFile *mmapFileImpl) readHeader() header {
	var result header
	buff := bytes.NewBuffer(mFile.memmap[:_HeaderSize])
	result.sanity = make([]byte, 15)

	binary.Read(buff, binary.BigEndian, &result.sanity)
	binary.Read(buff, binary.BigEndian, &result.ver)
	binary.Read(buff, binary.BigEndian, &result.mSize)

	return result
}

func (mFile *mmapFileImpl) sanityCheck(h header) bool {
	return true
}

func (mFile *mmapFileImpl) align(offset int) int {
	const alignment = 16
	rem := offset % alignment
	return offset + alignment - rem
}

/* Constructors */

// NewFile creates a new memory mapped file
func NewFile(fName string) (fs.File, error) {
	var err error

	result := new(mmapFileImpl)
	result.name = fName

	// Create/Open file
	result.file, err = os.OpenFile(fName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, errors.New("Could not open file for reading")
	}

	// Check to see if new
	info, _ := os.Stat(fName)
	result.mapSize = math.MaxInt(_InitialSize, int(info.Size()))

	if info.Size() == 0 {
		result.newFile = true
		result.file.Truncate(int64(result.mapSize))
	} else {
		result.newFile = false
	}

	// Map file to memory
	result.memmap, err = mmap.Map(result.file, mmap.RDWR, 0)

	// Validate
	if err != nil {
		return nil, err
	}

	result.lock = &sync.Mutex{}

	if !result.newFile {

		head := result.readHeader()

		common.Assert(bytes.Equal(head.sanity, _Sanity), "Sanity check failed '"+string(head.sanity)+"'")
		common.Assert(head.ver == _Version, "Versions do not match: ", head.ver, " vs. ", _Version)
		common.Assert(int(head.mSize) == result.mapSize, fName+": Sizes do not match: ", head.mSize, result.mapSize)

	} else {
		result.writeHeader()
	}
	return result, nil
}
