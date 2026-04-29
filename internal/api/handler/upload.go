package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/errors"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	cfg := config.C.XFileStorage
	if cfg.BaseURL == "" || cfg.APIKey == "" {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, "xfile_storage not configured"))
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, "file is required"))
		return
	}
	defer file.Close()

	path := c.PostForm("path")
	if path == "" {
		path = fmt.Sprintf("nexus/playground/%s/", time.Now().Format("2006-01-02"))
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	if _, err = io.Copy(part, file); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	writer.WriteField("path", path)
	writer.Close()

	req, _ := http.NewRequest("POST", cfg.BaseURL+"/api/v1/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Api-Key", cfg.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	r, err := client.Do(req)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, "upload failed: "+err.Error()))
		return
	}
	defer r.Body.Close()

	var result map[string]any
	json.NewDecoder(r.Body).Decode(&result)

	if r.StatusCode != http.StatusOK {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, "xfile_storage returned "+r.Status))
		return
	}

	resp.Success(c, result["data"])
}
