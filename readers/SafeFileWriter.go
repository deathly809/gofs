//
//	TODO: Think about this a lot.  Do we want to have the FileSystem
//	create and return a SafeWriter?
//

package readers

import (
	"io"

	"github.com/deathly809/gorapidstash/fs"
)

type fileWriter struct {
	file fs.File
}

func (writer *fileWriter) Write(data []byte) (written int, err error) {
	return writer.file.Write(data)
}

// NewSafeWriter takes in a File object and returns a writer that
// allows users to Write to the file
func NewSafeWriter(f File) io.Writer {
	result := new(fileWriter)
	result.file = f
	return result
}
