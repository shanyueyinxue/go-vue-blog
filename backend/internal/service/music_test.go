package service

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"blog/internal/app"
	"blog/internal/dto"
	"blog/internal/model"

	"github.com/stretchr/testify/require"
)

// tinyMP3 最小 ID3 头（http.DetectContentType 识别为 audio/mpeg）
var tinyMP3 = []byte("ID3\x04\x00\x00\x00\x00\x00\x00 audio-mpeg-stub")

// uploadTestFile 复用 FileService.Upload 上传测试文件，返回入库记录
func uploadTestFile(t *testing.T, a *app.App, name string, content []byte) *model.File {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile("file", name)
	require.NoError(t, err)
	_, err = fw.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	_, header, err := req.FormFile("file")
	require.NoError(t, err)

	file, err := NewFileService(New(a)).Upload(header, "", 1)
	require.Nil(t, err)
	return file
}

// TestUploadAudioMime 音频文件通过白名单校验，且未指定目录时自动存入 audios/
func TestUploadAudioMime(t *testing.T) {
	a := newTestApp(t)
	file := uploadTestFile(t, a, "song.mp3", tinyMP3)
	require.Equal(t, "audio/mpeg", file.MimeType)
	require.True(t, bytes.HasPrefix([]byte(file.Path), []byte("audios/")), "音频应存到 audios/，实际: %s", file.Path)
}

// TestMusicCreateListDelete 创建音乐 → 公开列表可见（带完整 URL）→ 删除时同步清理 files 记录与存储对象
func TestMusicCreateListDelete(t *testing.T) {
	a := newTestApp(t)
	musicSvc := NewMusicService(New(a))

	// 禁用状态音乐不入公开列表
	fileA := uploadTestFile(t, a, "a.mp3", tinyMP3)
	musicA, err := musicSvc.CreateMusic(&dto.MusicRequest{Name: "隐藏曲目", FileID: uintPtr(fileA.ID), Status: intPtr(0)})
	require.Nil(t, err)
	require.NotNil(t, musicA)

	fileB := uploadTestFile(t, a, "b.mp3", tinyMP3)
	musicB, err := musicSvc.CreateMusic(&dto.MusicRequest{Name: "公开曲目", Artist: strPtr("歌手"), FileID: uintPtr(fileB.ID)})
	require.Nil(t, err)
	require.NotNil(t, musicB)

	// 公开列表：仅启用曲目，且 URL 已拼装（浏览器直连存储）
	public, err := musicSvc.ListMusics()
	require.Nil(t, err)
	require.Len(t, public, 1)
	require.Equal(t, "公开曲目", public[0].Name)
	require.Equal(t, "歌手", public[0].Artist)
	require.Contains(t, public[0].URL, fileB.Path)

	// 管理端列表含禁用曲目（模拟 handler 先 Normalize 分页参数）
	adminQuery := dto.MusicQuery{}
	adminQuery.Normalize()
	adminList, total, err := musicSvc.ListAdminMusics(&adminQuery)
	require.Nil(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, adminList, 2)

	// 关键字搜索
	searchQuery := dto.MusicQuery{Keyword: "公开"}
	searchQuery.Normalize()
	searched, total, err := musicSvc.ListAdminMusics(&searchQuery)
	require.Nil(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "公开曲目", searched[0].Name)

	// 删除公开曲目：musics 软删除 + files 物理删除 + 存储对象清理
	require.Nil(t, musicSvc.DeleteMusic(musicB.ID))

	var musicCount int64
	require.NoError(t, a.DB.Model(&model.Music{}).Where("id = ?", musicB.ID).Count(&musicCount).Error)
	require.Zero(t, musicCount, "音乐记录应已删除")

	var fileCount int64
	require.NoError(t, a.DB.Model(&model.File{}).Where("id = ?", fileB.ID).Count(&fileCount).Error)
	require.Zero(t, fileCount, "关联文件记录应已删除")

	_, getErrB := a.Storage.GetFile(context.Background(), fileB.Path)
	require.Error(t, getErrB, "音频存储对象应已清理")

	// 未删除的音乐不受影响（注意关闭句柄，避免 Windows 下文件被锁定）
	rc, getErr := a.Storage.GetFile(context.Background(), fileA.Path)
	require.NoError(t, getErr)
	if rc != nil {
		require.NoError(t, rc.Close())
	}
}

// TestMusicUpdateReplaceFile 更换 file_id 时：旧文件记录与存储对象被清理，新文件保留
func TestMusicUpdateReplaceFile(t *testing.T) {
	a := newTestApp(t)
	musicSvc := NewMusicService(New(a))

	oldFile := uploadTestFile(t, a, "old.mp3", tinyMP3)
	newFile := uploadTestFile(t, a, "new.mp3", tinyMP3)

	music, err := musicSvc.CreateMusic(&dto.MusicRequest{Name: "测试曲目", FileID: uintPtr(oldFile.ID)})
	require.Nil(t, err)

	updated, err := musicSvc.UpdateMusic(music.ID, &dto.MusicRequest{Name: "测试曲目", FileID: uintPtr(newFile.ID)})
	require.Nil(t, err)
	require.Equal(t, newFile.ID, updated.FileID)

	// 旧文件记录已物理删除，存储对象已清理
	var oldCount int64
	require.NoError(t, a.DB.Model(&model.File{}).Where("id = ?", oldFile.ID).Count(&oldCount).Error)
	require.Zero(t, oldCount, "被替换的旧文件记录应已删除")
	_, getErrOld := a.Storage.GetFile(context.Background(), oldFile.Path)
	require.Error(t, getErrOld, "旧音频存储对象应已清理")

	// 新文件保留（注意关闭句柄，避免 Windows 下文件被锁定）
	rc, getErr := a.Storage.GetFile(context.Background(), newFile.Path)
	require.NoError(t, getErr)
	if rc != nil {
		require.NoError(t, rc.Close())
	}
}

// TestMusicCreateValidations 创建校验：歌名必填、音频文件必选、file_id 必须是音频
func TestMusicCreateValidations(t *testing.T) {
	a := newTestApp(t)
	musicSvc := NewMusicService(New(a))

	// 歌名为空
	_, err := musicSvc.CreateMusic(&dto.MusicRequest{Name: "  ", FileID: uintPtr(1)})
	require.NotNil(t, err)

	// 未选音频文件
	_, err = musicSvc.CreateMusic(&dto.MusicRequest{Name: "曲目"})
	require.NotNil(t, err)

	// file_id 指向不存在的文件
	_, err = musicSvc.CreateMusic(&dto.MusicRequest{Name: "曲目", FileID: uintPtr(99999)})
	require.NotNil(t, err)

	// file_id 指向非音频文件（上传一个 PDF）
	pdfFile := uploadTestFile(t, a, "doc.pdf", tinyPDF)
	_, err = musicSvc.CreateMusic(&dto.MusicRequest{Name: "曲目", FileID: uintPtr(pdfFile.ID)})
	require.NotNil(t, err, "非音频文件不应被关联为音乐")
}

// uintPtr uint 指针辅助
func uintPtr(u uint) *uint { return &u }
