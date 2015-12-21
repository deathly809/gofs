
package fs

import (
	"time"
)

// FileOffset is the relative offset in the file from
// a given  position
type FileOffset int

// Beginning
// Current
// End
const (
	Beginning FileOffset = iota
	Current
	End
)

// FileStats holds information related to a file
// TODO: Use this and fill it out
type FileStats interface {
	// Created returns the date and time the file was created
	Created() time.Time

	// LastModified returns the date and time the file was last modified
	LastModified() time.Time

	// Size returns the size of the file
	Size() int

	// Handles holds a list of current handles to the File
	// TODO: Too dangerous?  Who cares?  Security issues?
	Handles() []File
}

// File is the common interface all files will have
type File interface {
	// 	Close will close the flush all writes and close the file
	// 	any new writes will be ignored and all reads terminated
	Close() error

	//	Given a slice of data we will try to write this to the
	//	file.  Returns the number of bytes written and if an error
	//	occurs that is returned as well.
	//
	//	If an error is returned then the length written will be 0
	//
	Write(data []byte) (int, error)

	//	Read will attemp to read from the file and fill up the byte
	//	slice.  The amount of bytes written will be returned as the
	//	first return value.  If there is an error it will be returned
	//	as the second result, otherwise a nil will be returned instead.
	//
	//	If an error occurs the amount of data read will be 0
	//
	Read(data []byte) (int, error)

	//	Seek will try to move the file pointer to the position given
	//	from the offset provided.
	//
	//	If the resuting position is out of bounds it is clamped to
	//	either the beginning or the end of the file.
	//
	Seek(int, FileOffset)

	//	IsNew returns true if this file was created during opening.
	IsNew() bool

	// Name returns the name of the file.
	Name() string

	// Size returns the size of the file in bytes
	Size() int
}
