package http

import (
	"fmt"
	"io"
	"net/http"
)

type UploadedFile struct {
	TargetPath string
	Filename   string
	Content    []byte
}

func ParseUploadedFileRequest(r *http.Request) (*UploadedFile, error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, fmt.Errorf("parse multipart form: %w", err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("read form file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read uploaded file: %w", err)
	}

	return &UploadedFile{
		TargetPath: r.FormValue("targetPath"),
		Filename:   header.Filename,
		Content:    content,
	}, nil
}
