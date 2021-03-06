package log

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	golog "log"

	"github.com/stretchr/testify/assert"
)

func TestGoLogOutputGlobal(t *testing.T) {
	fmt.Println("\n\n--------------- TestGoLogOutputGlobal ---------------------")
	SetWriter(testNewConsoleWriter())
	golog.SetOutput(Outputer("golog", LevelInfo))
	golog.Print("hello", "golog")
}

func TestGoLogOutputNewLog(t *testing.T) {
	fmt.Println("\n\n--------------- TestGoLogOutputNewLog ---------------------")
	log := NewLog()
	log.SetWriter(testNewConsoleWriter())
	golog.SetOutput(log.Outputer("std", LevelInfo))
	golog.Print("hello", "golog")
}

func TestGoLogFileCallerGlobal(t *testing.T) {
	path := "TestGoLogFileCallerGlobal/filetest"
	dir := filepath.Dir(path)
	os.RemoveAll(dir)
	defer os.RemoveAll(dir)

	SetFormatter(NewTextFormatter("%l %S:%L %F() - %m"))
	SetWriter(&FileWriter{Path: path})
	golog.SetFlags(0)
	golog.SetOutput(Outputer("golog", LevelInfo, 3))
	file, line, ffun := testGetCaller(1)
	golog.Print("hello", "golog")
	Close()

	bs, _ := ioutil.ReadFile(path + ".log")
	assert.Equal(t, fmt.Sprintf("INFO %s:%d %s() - hellogolog\n", file, line, ffun), string(bs))
}

func TestGoLogFileCallerNewLog(t *testing.T) {
	path := "TestGoLogFileCallerNewLog/filetest"
	dir := filepath.Dir(path)
	os.RemoveAll(dir)
	defer os.RemoveAll(dir)

	log := NewLog()
	log.SetFormatter(NewTextFormatter("%l %S:%L %F() - %m"))
	log.SetWriter(&FileWriter{Path: path})
	golog.SetFlags(0)
	golog.SetOutput(log.Outputer("std", LevelInfo, 3))
	file, line, ffun := testGetCaller(1)
	golog.Print("hello", "golog")
	log.Close()

	bs, _ := ioutil.ReadFile(path + ".log")
	assert.Equal(t, fmt.Sprintf("INFO %s:%d %s() - hellogolog\n", file, line, ffun), string(bs))
}

func TestIoWriterFileCallerGlobal(t *testing.T) {
	path := "TestIoWriterFileCallerGlobal/filetest"
	dir := filepath.Dir(path)
	os.RemoveAll(dir)
	defer os.RemoveAll(dir)

	SetFormatter(NewTextFormatter("%l %S:%L %F() - %m%n"))
	SetWriter(&FileWriter{Path: path})

	iow := Outputer("iow", LevelInfo)
	file, line, ffun := testGetCaller(1)
	iow.Write(([]byte)("hello writer"))
	Close()

	bs, _ := ioutil.ReadFile(path + ".log")
	assert.Equal(t, fmt.Sprintf("INFO %s:%d %s() - hello writer"+eol, file, line, ffun), string(bs))
}

func TestIoWriterFileCallerNewLog(t *testing.T) {
	path := "TestIoWriterFileCallerNewLog/filetest"
	dir := filepath.Dir(path)
	os.RemoveAll(dir)
	defer os.RemoveAll(dir)

	log := NewLog()
	log.SetFormatter(NewTextFormatter("%l %S:%L %F() - %m%n"))
	log.SetWriter(&FileWriter{Path: path})

	iow := log.Outputer("iow", LevelInfo)
	file, line, ffun := testGetCaller(1)
	iow.Write(([]byte)("hello writer"))
	log.Close()

	bs, _ := ioutil.ReadFile(path + ".log")
	assert.Equal(t, fmt.Sprintf("INFO %s:%d %s() - hello writer"+eol, file, line, ffun), string(bs))
}
