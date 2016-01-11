package concrete

import (
	"errors"
	"time"

	"github.com/deathly809/gofs"
	"github.com/deathly809/gomath"
)

const (
	_Open   = iota
	_Closed = iota
)

// Meta-data about each file
type fileInfo struct {
	name         string
	size         int
	first        int
	last         int
	created      time.Time
	lastModified time.Time
}

// Logical information about an open file
type file struct {
	fs     *fileSystemImpl
	pos    int
	curr   fileNode
	fInfo  fileInfo
	isnew  bool
	status int
}

func (f *file) Close() error {
	f.status = _Closed
	return nil
}

func (f *file) positionsInSameBlock(posA, posB int) bool {
	return (posA / _BlockSize) == (posB / _BlockSize)
}

func (f *file) moveDown(bytes int) {
	finalPos := f.pos - bytes

	if f.positionOutOfBounds(finalPos) {

		finalPos = -finalPos
		blocksNeeded := int((finalPos + _BlockSize - 1) / _BlockSize)
		head, tail := f.fs.allocateBlocks(blocksNeeded)

		f.Seek(0, gofs.Beginning)

		tail.next = f.curr.id
		f.curr.prev = tail.id

		f.fInfo.first = head.id
		f.fInfo.size += finalPos

		f.curr = head
		f.pos = finalPos
	} else {

		currMove := f.pos % _BlockSize
		f.pos -= currMove
		bytes -= currMove

		for bytes >= _BlockSize {
			f.curr = f.fs.getBlock(f.curr.prev)
			currMove := gomath.Minint(_BlockSize, bytes)
			f.pos -= currMove
			bytes -= currMove
		}
	}
}

func (f *file) moveUp(bytes int) {
	finalPos := f.pos + bytes

	if f.positionOutOfBounds(finalPos) {
		bytesToGrowBy := finalPos - f.Size()
		blocksNeeded := (bytesToGrowBy + _BlockSize - 1) / _BlockSize
		head, tail := f.fs.allocateBlocks(blocksNeeded)

		f.Seek(0, gofs.End)
		head.prev = f.curr.id
		head.prev = f.curr.id

		f.fInfo.last = tail.id
		f.fInfo.size += bytesToGrowBy

		f.curr = tail
		f.pos = finalPos
	} else {

		currMove := f.pos % _BlockSize
		f.pos += currMove
		bytes -= currMove

		for bytes >= _BlockSize {
			f.curr = f.fs.getBlock(f.curr.prev)
			currMove := gomath.Minint(_BlockSize, bytes)
			f.pos += currMove
			bytes -= currMove
		}
	}
}

func (f *file) positionOutOfBounds(pos int) bool {
	return pos >= f.fInfo.size || pos < 0
}

func (f *file) growBy(size int) {
	currPos := f.pos
	endPos := f.fInfo.size + size
	f.Seek(-endPos, gofs.End)
	f.Seek(currPos, gofs.Beginning)
}

func (f *file) singleBlockWriteAtPos(data []byte) {
	offset := f.pos % _BlockSize
	bytesToWrite := gomath.Minint(_BlockSize-offset, len(data))

	copy(f.curr.data[offset:bytesToWrite], data)
}

func (f *file) singleBlockReadFromPos(data []byte) {
	offset := f.pos % _BlockSize
	bytesToRead := gomath.Minint(_BlockSize-offset, len(data))

	data = data[:bytesToRead]
	copy(data, f.curr.data[offset:bytesToRead])
}

func (f *file) Write(data []byte) (bytesWritten int, err error) {
	if f.status == _Closed {
		bytesWritten, err = 0, errors.New("Cannot write to a closed file")
	} else if data == nil {
		bytesWritten, err = 0, errors.New("Null pointer exception")
	} else {

		bytesWritten = len(data)
		finalPos := f.pos + len(data)

		if f.positionOutOfBounds(finalPos) {
			f.growBy(finalPos - f.Size())
		}

		bytesToWriteToBlock := _DataSize - (f.pos % _DataSize)
		for f.pos < finalPos {

			f.singleBlockWriteAtPos(data)
			f.Seek(bytesToWriteToBlock, gofs.Current)

			data = data[bytesToWriteToBlock:]
			bytesToWriteToBlock = gomath.Minint(_DataSize, len(data))

		}
	}

	return bytesWritten, err
}

// Read data into a given byte array
// If the array is null an error is returned
func (f *file) Read(data []byte) (bytesRead int, err error) {
	if f.status == _Closed {
		bytesRead, err = 0, errors.New("Cannot read from a closed file")
	} else if data == nil {
		bytesRead, err = 0, errors.New("Null pointer exception")
	} else {

		finalPos := f.pos + gomath.Minint(int(len(data)), f.fInfo.size-f.pos)

		bytesToReadFromBlock := _DataSize - (f.pos % _DataSize)
		for f.pos < finalPos {
			f.singleBlockReadFromPos(data)
			f.Seek(bytesToReadFromBlock, gofs.Current)

			data = data[bytesToReadFromBlock:]
			bytesToReadFromBlock = gomath.Minint(_DataSize, int(len(data)))
		}

		bytesRead, err = finalPos-f.pos, nil
	}
	return bytesRead, err
}

// Seek will move to a specific spot in the file.  If the
// spot is not within the file we expand the file to include
// it.
//
// If we seek before the beginning file we prepend zeros
// If we seek after the end of the file we append zeros
//
func (f *file) Seek(offset int, from gofs.FileOffset) {
	fSys := f.fs

	switch from {
	case gofs.Beginning:
		f.curr = fSys.getBlock(f.fInfo.first)
		f.pos = 0
	case gofs.End:
		f.curr = fSys.getBlock(f.fInfo.last)
		f.pos = f.fInfo.size
	}

	finalPos := f.pos + offset
	if !f.positionsInSameBlock(finalPos, f.pos) {
		if offset < 0 {
			f.moveDown(-offset)
		} else {
			f.moveUp(offset)
		}
	} else {
		f.pos = finalPos
	}
}

func (f *file) IsNew() bool {
	return f.isnew
}

func (f *file) Name() string {
	return f.fInfo.name
}

func (f *file) Size() int {
	return f.fInfo.size
}
