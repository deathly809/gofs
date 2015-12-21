package mmap

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/deathly809/gorapidstash/fs"
)

const (
	_LargeFile = 1024 * 1024
)

var testData = []byte("asdfgasdfgasdfgasdfgasdfgasdfg")
var testPath = filepath.Join(os.TempDir(), "testdata")

type wrapper struct {
	derp *testing.T
}

func init() {
	os.Remove(testPath)
	log.SetFlags(log.Ltime | log.LstdFlags | log.Lshortfile)
}

func TestCreate(t *testing.T) {

	file, err := NewFile(testPath)
	if err != nil {
		t.Error(err.Error())
		return
	}
	file.Close()

	info, err := os.Stat(testPath)
	if err != nil {
		t.Error("Could not create file")
		return
	}

	if info.Size() != int64(_InitialSize) {
		t.Error("Incorrect initial size")
		return
	}

	file.Close()
}

func TestWrite(t *testing.T) {

	file, err := NewFile(testPath)
	if err != nil {
		t.Error(err.Error())
		return
	}
	n, err := file.Write(testData)

	if n != len(testData) {
		t.Error("Did not write all data")
	}

	data := make([]byte, len(testData))
	file.Seek(0, fs.Beginning)
	n, err = file.Read(data)

	if n != len(testData) {
		t.Error("Did no read all data from file")
	}

	if err != nil {
		t.Error(err.Error())
	}

	same := bytes.Equal(testData, data)
	if !same {
		t.Error("Data not the same")
	}

	file.Close()

}

func TestRead(t *testing.T) {
	file, err := NewFile(testPath)

	if err != nil {
		t.Error(err.Error())
		return
	}

	data := make([]byte, len(testData))
	n, err := file.Read(data)

	if n != len(testData) {
		t.Error("Did not read all data from file")
	}
	if err != nil {
		t.Error(err.Error())
	}

	same := bytes.Equal(testData, data)
	if !same {
		t.Error("Data not the same")
	}

	file.Close()
}

func TestWrite_LargeFile(t *testing.T) {
	data := make([]byte, _LargeFile)
	for i := 0; i < _LargeFile; i++ {
		data[i] = byte(i % 256)
	}

	file, err := NewFile(testPath)

	if err != nil {
		t.Error(err.Error())
		return
	}

	n, err := file.Write(data)
	if n != _LargeFile {
		t.Error("Did not write all data to file")
	}

	if err != nil {
		t.Error(err.Error())
	}
	file.Close()
}

func TestRead_LargeFile(t *testing.T) {
	data := make([]byte, _LargeFile)
	for i := 0; i < _LargeFile; i++ {
		data[i] = byte(i % 256)
	}

	file, err := NewFile(testPath)

	if err != nil {
		t.Error(err.Error())
		return
	}
	test := make([]byte, _LargeFile)
	n, err := file.Read(test)

	if n != _LargeFile {
		t.Error("Did not read all data from file")
	}
	if err != nil {
		t.Error(err.Error())
		return
	}
	file.Close()

	if !bytes.Equal(data, test) {
		t.Error("Data not correct")
		return
	}
}

func TestTearDown(t *testing.T) {
	os.Remove(testPath)
	info, err := os.Stat(testPath)
	if err == nil {
		t.Error("File still exists: ", testPath)
	}

	if info != nil {
		t.Error("File still exists")
	}
}
