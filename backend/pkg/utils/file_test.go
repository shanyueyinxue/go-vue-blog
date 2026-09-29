package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitPathAndName(t *testing.T) {
	// 示例 1：绝对路径（Linux）
	d, f := SplitPathAndName("/home/user/docs/photo.jpg")
	t.Logf("dir=%q, name=%q\n", d, f) // dir="/home/user/docs", name="photo.jpg"
	assert.Equal(t, "/home/user/docs", d)
	assert.Equal(t, "photo.jpg", f)

	// 示例 2：相对路径（包含分隔符）
	d, f = SplitPathAndName("./data/file.txt")
	t.Logf("dir=%q, name=%q\n", d, f) // dir="data", name="file.txt"
	assert.Equal(t, "data", d)
	assert.Equal(t, "file.txt", f)

	// 示例 2：相对路径（不包含分隔符）
	d, f = SplitPathAndName("data/file.txt")
	t.Logf("dir=%q, name=%q\n", d, f) // dir="data", name="file.txt"
	assert.Equal(t, "data", d)
	assert.Equal(t, "file.txt", f)

	// 示例 3：仅有文件名
	d, f = SplitPathAndName("readme.md")
	t.Logf("dir=%q, name=%q\n", d, f) // dir="", name="readme.md"
	assert.Equal(t, "", d)
	assert.Equal(t, "readme.md", f)

	// 示例 4：Windows 路径
	d, f = SplitPathAndName(`C:\Program Files\app.exe`)
	t.Logf("dir=%q, name=%q\n", d, f) // dir="C:/Program Files", name="app.exe"
	assert.Equal(t, `C:/Program Files`, d)
	assert.Equal(t, "app.exe", f)

	// 示例 5：根目录下的文件
	d, f = SplitPathAndName("/file.txt")
	t.Logf("dir=%q, name=%q\n", d, f) // dir="/", name="file.txt"
	assert.Equal(t, "/", d)
	assert.Equal(t, "file.txt", f)
}
