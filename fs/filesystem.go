package fs

import "io"

// FileSystem is an interface into your brain
type FileSystem interface {

	//	GetSafeWriter returns a writer for the file provided
	//
	//	All writes to the file are atomic, if any reads are
	//	in progress during a write the writer will wait for
	//  those in progress to complete.  All further reads
	//	will be blocked until a write has completed
	//

	GetSafeWriter(File) io.Writer

	//	GetSafeReader returns a reader for the file provided.
	//
	//	A read will result in a consistent value.  If at the
	//  time of a read a write is in progress the read blocks
	//  until the write has continued.  It will then attempt
	//  another read.
	//
	GetSafeReader(File) io.Reader

	//	Shutdown safely closes the file system
	//
	//  All reads and writes in progress will finish and
	//	further I/O will result in an error.
	//
	//	Any future calls to the FileSystem object will
	//	have no effect.
	Shutdown()

	//	Lock provides exclusive access to a file.
	//
	//	When Lock is called on a file any reads or writes
	//	performed by a thread which does not have the lock
	//	is blocked.  The thread with the lock may perform
	//	any operations it wishes.
	//
	//	Operations by any thread spawned by the locking
	//	thread is undefined.
	//
	Lock(File)

	//	Unlock release the lock on the provided file
	//
	//	If the calling thread does not have the lock nothing
	//  happens.
	//
	Unlock(File)

	//	Open locates and returns a file in the file system
	//
	//  If the file does not exist it is created with 0 length
	//
	//	If an error occurs nil is returned
	//
	Open(string) File

	//	Exists returns true if the file exists in the
	//	filesystem, otherwise it return false
	Exists(string) bool

	//	Delete removes an existing file from the filesystem
	//
	//	If the file does not exist nothing happens
	//
	Delete(string)
}
