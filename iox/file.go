package iox

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DirExists check if the directory dir exists
func DirExists(dir string) error {
	fi, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%q is not a directory", dir)
	}
	return nil
}

// FileExists check if the file exists
func FileExists(file string) error {
	fi, err := os.Stat(file)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("%q is directory", file)
	}
	return nil
}

// CopyFile copy src file to des file
func CopyFile(src string, dst string) error {
	ss, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !ss.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	dd := filepath.Dir(dst)
	os.MkdirAll(dd, ss.Mode().Perm())

	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, ss.Mode().Perm())
	if err != nil {
		return err
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	return err
}

// FileReader a file reader
type FileReader struct {
	Path string
	file *os.File
}

// Read implements io.Reader
func (fr *FileReader) Read(p []byte) (n int, err error) {
	if fr.file == nil {
		file, err := os.Open(fr.Path)
		if err != nil {
			return 0, err
		}
		fr.file = file
	}
	return fr.file.Read(p)
}

// Close implements io.Close
func (fr *FileReader) Close() error {
	if fr.file == nil {
		return nil
	}

	err := fr.file.Close()
	fr.file = nil
	return err
}
