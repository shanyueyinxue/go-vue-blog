package utils

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidFilePath(t *testing.T) {
	if runtime.GOOS != "windows" {
		// Unix-like 系统测试
		t.Log("Unix-like 系统测试")
		assert.False(t, IsValidFilePath(""))
		assert.True(t, IsValidFilePath("/home/user/file.txt"))
		assert.True(t, IsValidFilePath("relative/path/to/file"))
		assert.False(t, IsValidFilePath("file\x00name"))
		assert.False(t, IsValidFilePath("//home//user"))
	}
	if runtime.GOOS == "windows" {
		// Windows 系统测试
		t.Log("Windows 系统测试")
		assert.True(t, IsValidFilePath(`C:\Windows\System32`))
		assert.False(t, IsValidFilePath(`C:file.txt`))
		assert.False(t, IsValidFilePath(`file<.txt`))
		assert.False(t, IsValidFilePath(`CON`))
		assert.False(t, IsValidFilePath(`file `))
		assert.True(t, IsValidFilePath(`file`))
		assert.True(t, IsValidFilePath(`file/123`))
		assert.True(t, IsValidFilePath(`file.txt`))
	}
}

func TestPathJoin(t *testing.T) {
	t.Log(filepath.Join("a", "b", "c"))       // "a\\b\\c" on Windows, "a/b/c" on Unix-like
	t.Log(filepath.Join("a", "b", "c", ""))   // "a\\b\\c" on Windows, "a/b/c" on Unix-like
	t.Log(filepath.Join("a", "b", "c", "."))  // "a\\b\\c" on Windows, "a/b/c" on Unix-like
	t.Log(filepath.Join("a", "b", "c", "..")) // "a\\b" on Windows, "a/b" on Unix-like

	t.Log(filepath.Join("a", "b/c"))
	t.Log(filepath.Join("a", "b\\c"))
}
