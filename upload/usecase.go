package upload

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const DefaultMaxImageSize = 10 << 20 // 10 MB

var (
	ErrNoImageFile          = errors.New("at least one image file is required")
	ErrImageTooLarge        = errors.New("image exceeds 10MB limit")
	ErrUnsupportedImageType = errors.New("unsupported image type")
	ErrUploaderUnavailable  = errors.New("upload service is not configured")
)

type File struct {
	Filename string
	Size     int64
	Open     func() (io.ReadSeekCloser, error)
}

type UploadedFile struct {
	FileName    string
	URL         string
	Size        int64
	ContentType string
}

type Uploader interface {
	Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) (string, error)
}

type Service interface {
	UploadImages(ctx context.Context, folder string, files []File) ([]UploadedFile, error)
}

type Usecase struct {
	uploader     Uploader
	maxImageSize int64
}

func NewUsecase(uploader Uploader) *Usecase {
	return &Usecase{
		uploader:     uploader,
		maxImageSize: DefaultMaxImageSize,
	}
}

func (uc *Usecase) UploadImages(ctx context.Context, folder string, files []File) ([]UploadedFile, error) {
	if uc.uploader == nil {
		return nil, ErrUploaderUnavailable
	}

	if len(files) == 0 {
		return nil, ErrNoImageFile
	}

	folder = strings.Trim(strings.TrimSpace(folder), "/")
	if folder == "" {
		folder = "images"
	}

	uploaded := make([]UploadedFile, 0, len(files))

	for i := range files {
		file, err := uc.uploadImage(ctx, folder, files[i])
		if err != nil {
			return nil, err
		}

		uploaded = append(uploaded, file)
	}

	return uploaded, nil
}

func (uc *Usecase) uploadImage(ctx context.Context, folder string, file File) (UploadedFile, error) {
	if file.Size > uc.maxImageSize {
		return UploadedFile{}, ErrImageTooLarge
	}

	if file.Open == nil {
		return UploadedFile{}, fmt.Errorf("open uploaded file: open function is nil")
	}

	reader, err := file.Open()
	if err != nil {
		return UploadedFile{}, fmt.Errorf("open uploaded file: %w", err)
	}
	defer reader.Close()

	contentType, err := detectImageContentType(reader)
	if err != nil {
		return UploadedFile{}, err
	}

	objectKey := buildImageObjectKey(file.Filename, contentType, folder)

	url, err := uc.uploader.Upload(ctx, objectKey, reader, file.Size, contentType)
	if err != nil {
		return UploadedFile{}, fmt.Errorf("upload image to storage: %w", err)
	}

	return UploadedFile{
		FileName:    file.Filename,
		URL:         url,
		Size:        file.Size,
		ContentType: contentType,
	}, nil
}

func detectImageContentType(reader io.ReadSeeker) (string, error) {
	buf := make([]byte, 512)

	n, err := reader.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read file header: %w", err)
	}

	if _, err = reader.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("reset file reader: %w", err)
	}

	contentType := http.DetectContentType(buf[:n])
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return contentType, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedImageType, contentType)
	}
}

func buildImageObjectKey(filename, contentType, folder string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = contentTypeToExtension(contentType)
	}

	if ext == "" {
		ext = ".bin"
	}

	return folder + "/" + time.Now().UTC().Format("20060102") + "/" + randomHex(16) + ext
}

func contentTypeToExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

func randomHex(bytesCount int) string {
	b := make([]byte, bytesCount)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(b)
}
