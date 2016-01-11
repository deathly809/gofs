package concrete

import (
	"errors"
	"os"

	"github.com/deathly809/gofs"
)

/*
   Concrete Filesystem Setup


       The actual filesystem consists of two files: name file, and data file.

       The name file contains a fixed length header.

       [SIGNATURE:NUMBER_OF_FILES]

       where the size of each in bytes is:

       [8:8]

       After the header there are a fixed number of files to read as specified
       by the header.  Each file has an entry of the form:

       [NAME : FIRST : LAST : SIZE]

       where the size of each in bytes is:

       [256:8:8:8]

*/

func readNames(f *os.File) map[string]int {
	return nil
}

// Open creates the default filesystem
func Open(directory, name string) (gofs.FileSystem, error) {

	result := &int{}
	result.fsDirectory = directory
	result.fsName = name

	err := result.init()

	if err != nil {
		return nil, errors.New("Could not open filesystem: " + err.Error())
	}

	return result, nil
}
