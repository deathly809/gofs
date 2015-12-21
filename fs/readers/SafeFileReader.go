//
//	TODO: Think about this a lot.  Do we want to have the FileSystem
//	create and return a SafeReader?
//

package readers

import "io"

type fileReader struct {
	file File
}

func (reader *fileReader) Read(p []byte) (int, error) {
	return reader.file.Read(p)

}

// NewSafeReader takes in a File object and returns a reader that
// allows users to write to the file
func NewSafeReader(f File) io.Reader {
	result := new(fileReader)
	result.file = f
	return result
}
