package log

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileWriter implements Writer.
// It writes messages and rotate by file size limit, daily, hourly.
type FileWriter struct {
	Path      string    // Log file path name
	DirPerm   uint32    // Log dir permission
	FilePerm  uint32    // Log file permission
	MaxSplit  int       // Max split files
	MaxSize   int64     // Rotate at size
	MaxDays   int       // Max daily files
	MaxHours  int       // Max hourly files
	Gzip      bool      // Compress rotated log files
	SyncLevel Level     // Call File.Sync() if level <= SyncLevel
	Logfmt    Formatter // log formatter
	Logfil    Filter    // log filter

	dir      string
	prefix   string
	suffix   string
	file     *os.File
	fileSize int64
	fileNum  int
	openTime time.Time
	openDay  int
	openHour int
	bb       bytes.Buffer
}

// SetSyncLevel set the sync level
func (fw *FileWriter) SetSyncLevel(lvl string) {
	fw.SyncLevel = ParseLevel(lvl)
}

// SetFormat set the log formatter
func (fw *FileWriter) SetFormat(format string) {
	fw.Logfmt = NewLogFormatter(format)
}

// SetFilter set the log filter
func (fw *FileWriter) SetFilter(filter string) {
	fw.Logfil = NewLogFilter(filter)
}

// Write write logger message into file.
func (fw *FileWriter) Write(le *Event) {
	if fw.Logfil != nil && fw.Logfil.Reject(le) {
		return
	}

	lf := fw.Logfmt
	if lf == nil {
		lf = le.Logger.GetFormatter()
		if lf == nil {
			lf = TextFmtDefault
		}
	}

	fw.init()
	if fw.file == nil {
		return
	}

	if fw.fileSize > 0 && fw.needRotate(le) {
		fw.rotate(le.When)
	}

	// format msg
	fw.bb.Reset()
	lf.Write(&fw.bb, le)

	// write log
	n, err := fw.file.Write(fw.bb.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileWriter(%q) - Write(): %v\n", fw.Path, err)
	}
	fw.fileSize += int64(n)

	if le.Level <= fw.SyncLevel {
		err := fw.file.Sync()
		if err != nil {
			fmt.Fprintf(os.Stderr, "FileWriter(%q) - Sync(): %v\n", fw.Path, err)
		}
	}
}

// Flush flush file logger.
// there are no buffering messages in file logger in memory.
// flush file means sync file from disk.
func (fw *FileWriter) Flush() {
	if fw.file != nil {
		err := fw.file.Sync()
		if err != nil {
			fmt.Fprintf(os.Stderr, "FileWriter(%q) - Sync(): %v\n", fw.Path, err)
		}
	}
}

// Close close the file description, close file writer.
func (fw *FileWriter) Close() {
	if fw.file != nil {
		err := fw.file.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "FileWriter(%q) - Close(): %v\n", fw.Path, err)
		}
		fw.file = nil
	}
}

func (fw *FileWriter) init() {
	if fw.file != nil {
		return
	}

	// init dir, prefix, suffix
	if fw.prefix == "" {
		fw.dir, fw.prefix = filepath.Split(fw.Path)
		fw.suffix = filepath.Ext(fw.prefix)
		if fw.suffix == "" {
			fw.suffix = ".log"
			fw.Path += fw.suffix
		}
		fw.prefix = strings.TrimSuffix(fw.prefix, fw.suffix)
	}

	// init perm
	if fw.DirPerm == 0 {
		fw.DirPerm = 0770
	}
	if fw.FilePerm == 0 {
		fw.FilePerm = 0660
	}

	// create dirs
	os.MkdirAll(fw.dir, os.FileMode(fw.DirPerm))

	// Open the log file
	file, err := os.OpenFile(fw.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(fw.FilePerm))
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileWriter(%q) - OpenFile(): %v\n", fw.Path, err)
		return
	}

	// Make sure file perm is user set perm cause of `os.OpenFile` will obey umask
	os.Chmod(fw.Path, os.FileMode(fw.FilePerm))

	fi, err := file.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileWriter(%q) - Stat(): %v\n", fw.Path, err)
		return
	}

	// init file info
	fw.fileSize = fi.Size()
	fw.openTime = fi.ModTime()
	fw.openDay = fw.openTime.Day()
	fw.openHour = fw.openTime.Hour()

	fw.file = file
}

func (fw *FileWriter) needRotate(le *Event) bool {
	return (fw.MaxSize > 0 && fw.fileSize >= fw.MaxSize) ||
		(fw.MaxHours > 0 && fw.openHour != le.When.Hour()) ||
		(fw.MaxDays > 0 && fw.openDay != le.When.Day())
}

// DoRotate means it need to write file in new file.
// new file name like xx-20130101.log (daily) or xx-001.log (by line or size)
func (fw *FileWriter) rotate(tm time.Time) {
	path := "" // rotate file name

	date := ""
	if fw.MaxHours > 0 {
		date = fw.openTime.Format("-2006010215")
		if fw.openHour != tm.Hour() {
			fw.fileNum = 0
		}
	} else if fw.MaxDays > 0 {
		date = fw.openTime.Format("-20060102")
		if fw.openDay != tm.Day() {
			fw.fileNum = 0
		}
	}

	pre := filepath.Join(fw.dir, fw.prefix) + date
	if fw.MaxSize > 0 {
		// get splited next file name
		path = fw.nextFile(pre)
	} else {
		path = pre + fw.suffix
		_, err := os.Stat(path)
		if err == nil {
			// timely rotate file exists (normally impossible)
			// find next split file name
			path = fw.nextFile(pre)
		}
	}

	// close file before rename
	err := fw.file.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileWriter(%q) - Close(): %v\n", fw.Path, err)
		return
	}
	fw.file = nil

	// Rename the file to its new found name
	// even if occurs error,we MUST guarantee to  restart new logger
	err = os.Rename(fw.Path, path)
	if err == nil {
		if fw.Gzip {
			go fw.compressFile(path)
		}
	} else {
		fmt.Fprintf(os.Stderr, "FileWriter(%q) - Rename(->%q): %v\n", fw.Path, path, err)
	}

	// Open file again
	fw.init()

	// delete outdated rotated files
	if fw.MaxHours > 0 || fw.MaxDays > 0 {
		go fw.deleteOutdatedFiles()
	}
}

func (fw *FileWriter) nextFile(pre string) string {
	var path string
	for fw.fileNum++; ; fw.fileNum++ {
		path = pre + fmt.Sprintf("-%03d", fw.fileNum) + fw.suffix
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			if fw.Gzip {
				p := path + ".gz"
				_, err = os.Stat(p)
				if os.IsNotExist(err) {
					break
				}
			} else {
				break
			}
		}
	}

	if fw.MaxSplit > 0 && fw.fileNum > fw.MaxSplit {
		// remove old splited files
		for i := fw.fileNum - fw.MaxSplit; i > 0; i-- {
			p := pre + fmt.Sprintf("-%03d", i) + fw.suffix
			err := os.Remove(p)
			if os.IsNotExist(err) {
				if fw.Gzip {
					pg := path + ".gz"
					err = os.Remove(p)
					if os.IsNotExist(err) {
						break
					} else if err != nil {
						fmt.Fprintf(os.Stderr, "FileWriter(%q) - Remove(%q): %v\n", fw.Path, pg, err)
					}
				} else {
					break
				}
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "FileWriter(%q) - Remove(%q): %v\n", fw.Path, p, err)
			}
		}
	}
	return path
}

func (fw *FileWriter) compressFile(src string) {
	dst := src + ".gz"

	f, err := os.Open(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileWriter(%q) - Open(%q): %v\n", fw.Path, src, err)
		return
	}
	defer f.Close()

	// If this file already exists, we presume it was created by
	// a previous attempt to compress the log file.
	gzf, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(fw.FilePerm))
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileWriter(%q) - OpenFile(%q): %v\n", fw.Path, dst, err)
		return
	}
	defer gzf.Close()

	gz := gzip.NewWriter(gzf)

	if _, err := io.Copy(gz, f); err != nil {
		fmt.Fprintf(os.Stderr, "FileWriter(%q) - gzip(%q): %v\n", fw.Path, dst, err)
		return
	}
	if err := gz.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "FileWriter(%q) - gzip.Close(%q): %v\n", fw.Path, dst, err)
		return
	}
	if err := gzf.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "FileWriter(%q) - Close(%q): %v\n", fw.Path, dst, err)
		return
	}

	f.Close()
	if err := os.Remove(src); err != nil {
		fmt.Fprintf(os.Stderr, "FileWriter(%q) - Remove(%q): %v\n", fw.Path, src, err)
	}
}

func (fw *FileWriter) deleteOutdatedFiles() {
	var due time.Time
	if fw.MaxHours > 0 {
		due = time.Now().Add(-1 * time.Hour * time.Duration(fw.MaxHours))
	} else {
		due = time.Now().Add(-24 * time.Hour * time.Duration(fw.MaxDays))
	}

	f, err := os.Open(fw.dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileWriter(%q) - Open(%q): %v\n", fw.Path, fw.dir, err)
		return
	}
	defer f.Close()

	fis, err := f.Readdir(-1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileWriter(%q) - Readdir(%q): %v\n", fw.Path, fw.dir, err)
		return
	}

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		if fi.ModTime().Before(due) {
			name := filepath.Base(fi.Name())
			if strings.HasPrefix(name, fw.prefix) {
				path := filepath.Join(fw.dir, fi.Name())
				if err := os.Remove(path); err != nil {
					fmt.Fprintf(os.Stderr, "FileWriter(%q) - Remove(%q): %v\n", fw.Path, path, err)
				}
			}
		}
	}
}

func init() {
	RegisterWriter("file", func() Writer {
		return &FileWriter{}
	})
}
