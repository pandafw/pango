package main

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathMatch1(t *testing.T) {
	b, err := filepath.Match("dir/*.txt", "dir/a.txt")
	if !b || err != nil {
		t.Errorf(`filepath.Match("dir/*.txt", "dir/a.txt") = %v, %v`, b, err)
	}
}

func TestPathMatch2(t *testing.T) {
	b, err := filepath.Match("dir/**/*.txt", "dir/a.txt")
	if b || err != nil {
		t.Errorf(`filepath.Match("dir/**/*.txt", "dir/a.txt") = %v, %v`, b, err)
	}
}

func TestPathMatch3(t *testing.T) {
	b, err := filepath.Match("dir/**/*.txt", "dir/3/a.txt")
	if !b || err != nil {
		t.Errorf(`filepath.Match("dir/**/*.txt", "dir/3/a.txt") = %v, %v`, b, err)
	}
}

func TestPathMatch4(t *testing.T) {
	b, err := filepath.Match("dir/**/*.txt", "dir/3/5/a.txt")

	// why??
	if runtime.GOOS == "windows" {
		if !b || err != nil {
			t.Errorf(`filepath.Match("dir/**/*.txt", "dir/3/5/a.txt") = %v, %v`, b, err)
		}
	} else {
		if b || err != nil {
			t.Errorf(`filepath.Match("dir/**/*.txt", "dir/3/5/a.txt") = %v, %v`, b, err)
		}
	}
}

func TestPathMatch5(t *testing.T) {
	b, err := filepath.Match("**/*.txt", "a.txt")
	assert.Nil(t, err)
	assert.False(t, b)
}

func TestPathMatch6(t *testing.T) {
	b, err := filepath.Match("**/*.txt", "a/a.txt")
	assert.Nil(t, err)
	assert.True(t, b)

	b, err = filepath.Match("**/*.txt", "a\\a.txt")
	assert.Nil(t, err)
	assert.False(t, b)

	if runtime.GOOS == "windows" {
		b, err = filepath.Match("**\\*.txt", "a\\a.txt")
		assert.Nil(t, err)
		assert.True(t, b)
	}
}
