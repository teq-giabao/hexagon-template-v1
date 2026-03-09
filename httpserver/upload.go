package httpserver

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"hexagon/upload"

	"github.com/labstack/echo/v4"
)

func (s *Server) handleUploadImages(c echo.Context, folder string) error {
	if s.UploadService == nil {
		return respondError(c, http.StatusNotImplemented, "upload service is not configured", "")
	}

	form, err := c.MultipartForm()
	if err != nil {
		return respondError(c, http.StatusBadRequest, "invalid multipart/form-data request", err.Error())
	}

	files := toUploadFiles(collectImageFiles(form))

	uploaded, err := s.UploadService.UploadImages(c.Request().Context(), folder, files)
	if err != nil {
		switch {
		case errors.Is(err, upload.ErrNoImageFile):
			return respondError(c, http.StatusBadRequest, upload.ErrNoImageFile.Error(), "")
		case errors.Is(err, upload.ErrImageTooLarge), errors.Is(err, upload.ErrUnsupportedImageType):
			return respondError(c, http.StatusBadRequest, "invalid image file", err.Error())
		case errors.Is(err, upload.ErrUploaderUnavailable):
			return respondError(c, http.StatusNotImplemented, "upload service is not configured", "")
		default:
			return respondError(c, http.StatusInternalServerError, "failed to upload image", err.Error())
		}
	}

	result := make([]UploadedImageResponse, len(uploaded))
	for i := range uploaded {
		result[i] = UploadedImageResponse{
			FileName:    uploaded[i].FileName,
			URL:         uploaded[i].URL,
			Size:        uploaded[i].Size,
			ContentType: uploaded[i].ContentType,
		}
	}

	return respondCreated(c, APIDataResult{Data: UploadImagesResponse{Files: result}})
}

func collectImageFiles(form *multipart.Form) []*multipart.FileHeader {
	if form == nil {
		return nil
	}

	files := make([]*multipart.FileHeader, 0, len(form.File["image"])+len(form.File["images"]))
	files = append(files, form.File["image"]...)
	files = append(files, form.File["images"]...)

	return files
}

func toUploadFiles(headers []*multipart.FileHeader) []upload.File {
	files := make([]upload.File, 0, len(headers))

	for i := range headers {
		header := headers[i]
		files = append(files, upload.File{
			Filename: header.Filename,
			Size:     header.Size,
			Open: func() (io.ReadSeekCloser, error) {
				return header.Open()
			},
		})
	}

	return files
}
