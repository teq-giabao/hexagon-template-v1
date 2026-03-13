package upload

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type readSeekCloser struct {
	*bytes.Reader
}

func (r *readSeekCloser) Close() error { return nil }

func newReadSeekCloser(b []byte) io.ReadSeekCloser {
	return &readSeekCloser{Reader: bytes.NewReader(b)}
}

type fakeUploader struct {
	lastKey         string
	lastContentType string
	lastSize        int64
	lastBody        []byte
	url             string
	err             error
}

func (f *fakeUploader) Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) (string, error) {
	f.lastKey = key
	f.lastContentType = contentType
	f.lastSize = size
	f.lastBody, _ = io.ReadAll(body)
	return f.url, f.err
}

func TestUsecase_UploadImages_UploaderUnavailable(t *testing.T) {
	uc := NewUsecase(nil)
	_, err := uc.UploadImages(context.Background(), "", nil)
	assert.ErrorIs(t, err, ErrUploaderUnavailable)
}

func TestUsecase_UploadImages_NoFiles(t *testing.T) {
	uc := NewUsecase(&fakeUploader{})
	_, err := uc.UploadImages(context.Background(), "", nil)
	assert.ErrorIs(t, err, ErrNoImageFile)
}

func TestUsecase_UploadImages_FileTooLarge(t *testing.T) {
	uc := NewUsecase(&fakeUploader{})
	file := File{Filename: "a.jpg", Size: uc.maxImageSize + 1, Open: func() (io.ReadSeekCloser, error) {
		return newReadSeekCloser([]byte("data")), nil
	}}

	_, err := uc.UploadImages(context.Background(), "images", []File{file})
	assert.ErrorIs(t, err, ErrImageTooLarge)
}

func TestUsecase_UploadImages_OpenFuncNil(t *testing.T) {
	uc := NewUsecase(&fakeUploader{})
	file := File{Filename: "a.jpg", Size: 10}

	_, err := uc.UploadImages(context.Background(), "images", []File{file})
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "open function is nil"))
}

func TestDetectImageContentType_Unsupported(t *testing.T) {
	reader := bytes.NewReader([]byte("hello"))
	_, err := detectImageContentType(reader)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedImageType))
}

func TestBuildImageObjectKey(t *testing.T) {
	date := time.Now().UTC().Format("20060102")
	key := buildImageObjectKey("photo", "image/jpeg", "uploads")

	assert.True(t, strings.HasPrefix(key, "uploads/"+date+"/"))
	assert.True(t, strings.HasSuffix(key, ".jpg"))
}

func TestUsecase_UploadImages_Success(t *testing.T) {
	uploader := &fakeUploader{url: "https://cdn.example.com/uploads/x.jpg"}
	uc := NewUsecase(uploader)

	pngHeader := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	file := File{
		Filename: "image.png",
		Size:     int64(len(pngHeader)),
		Open: func() (io.ReadSeekCloser, error) {
			return newReadSeekCloser(pngHeader), nil
		},
	}

	files, err := uc.UploadImages(context.Background(), "avatars", []File{file})
	require.NoError(t, err)
	require.Len(t, files, 1)

	assert.Equal(t, "image.png", files[0].FileName)
	assert.Equal(t, uploader.url, files[0].URL)
	assert.Equal(t, "image/png", files[0].ContentType)
	assert.True(t, strings.HasPrefix(uploader.lastKey, "avatars/"))
	assert.Equal(t, int64(len(pngHeader)), uploader.lastSize)
	assert.Equal(t, pngHeader, uploader.lastBody)
}
