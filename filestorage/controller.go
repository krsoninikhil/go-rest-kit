package filestorage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/krsoninikhil/go-rest-kit/apperrors"
	"github.com/krsoninikhil/go-rest-kit/auth"
	"github.com/krsoninikhil/go-rest-kit/integrations/aws"
	"github.com/krsoninikhil/go-rest-kit/request"
)

// ControllerOpts configures optional controller behavior (e.g. allowed prefixes).
// If AllowedPrefixes is non-nil, Upload will reject any prefix not in the list.
// Prefixes like "public/profiles" are for public objects; others (e.g. "frameworks") are private.
type ControllerOpts struct {
	AllowedPrefixes []string
}

var errFileHeaderRequired = errors.New("file header is required")

// UploadResponse is the JSON response for a successful upload.
type UploadResponse struct {
	Path string `json:"path"`
	URL  string `json:"url,omitempty"`
}

// SignedURLResponse is the JSON response for a signed URL.
type SignedURLResponse struct {
	URL string `json:"url"`
}

// Controller exposes Gin handlers for upload and signed URL using AWS S3.
type Controller struct {
	s3              *aws.S3
	allowedPrefixes map[string]struct{} // nil = no validation
}

// NewController creates a filestorage controller from AWS config.
// If opts is non-nil and AllowedPrefixes is set, only those prefixes are accepted on Upload.
func NewController(cfg aws.Config, opts *ControllerOpts) (*Controller, error) {
	s3Client, err := aws.NewS3(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	c := &Controller{s3: s3Client}
	if opts != nil && len(opts.AllowedPrefixes) > 0 {
		c.allowedPrefixes = make(map[string]struct{})
		for _, p := range opts.AllowedPrefixes {
			c.allowedPrefixes[p] = struct{}{}
		}
	}
	return c, nil
}

// NewControllerWithS3 creates a controller with an existing S3 client (e.g. for testing).
func NewControllerWithS3(s3 *aws.S3) *Controller {
	return &Controller{s3: s3}
}

// Upload handles POST multipart form: file "file", query "prefix" (required).
// Public access is derived from prefix: prefixes starting with "public/" are treated as public (return permanent URL); others get a presigned URL.
func (c *Controller) Upload(ctx *gin.Context) {
	userID := auth.UserID(ctx)
	if userID == 0 {
		request.Respond(ctx, nil, apperrors.NewPermissionError("user_id"))
		return
	}
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		request.Respond(ctx, nil, apperrors.NewInvalidParamsError("filestorage", err))
		return
	}
	defer file.Close()
	if header == nil {
		request.Respond(ctx, nil, apperrors.NewInvalidParamsError("filestorage", errFileHeaderRequired))
		return
	}

	prefix := ctx.Query("prefix")
	public := strings.HasPrefix(prefix, "public/")

	if c.allowedPrefixes != nil {
		if _, ok := c.allowedPrefixes[prefix]; !ok {
			request.Respond(ctx, nil, apperrors.NewInvalidParamsError("filestorage", errors.New("prefix not allowed")))
			return
		}
	}

	key := buildKey(header.Filename, prefix, userID)
	body, err := io.ReadAll(file)
	if err != nil {
		request.Respond(ctx, nil, apperrors.NewServerError(err))
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if err := c.s3.PutObject(ctx.Request.Context(), key, bytes.NewReader(body), contentType, public); err != nil {
		request.Respond(ctx, nil, err)
		return
	}

	var url string
	if public {
		url = c.s3.PublicURL(key)
	} else {
		signed, err := c.s3.PresignGet(ctx.Request.Context(), key, time.Hour)
		if err == nil {
			url = signed
		}
	}
	request.Respond(ctx, &UploadResponse{Path: key, URL: url}, nil)
}

// SignedURL handles GET query path= and optional expiry= (duration string, e.g. "1h").
func (c *Controller) SignedURL(ctx *gin.Context) {
	userID := auth.UserID(ctx)
	if userID == 0 {
		request.Respond(ctx, nil, apperrors.NewPermissionError("user_id"))
		return
	}
	path := ctx.Query("path")
	if path == "" {
		request.Respond(ctx, nil, apperrors.NewInvalidParamsError("filestorage", errors.New("path is required")))
		return
	}
	expiryStr := ctx.Query("expiry")
	expiry := time.Hour
	if expiryStr != "" {
		if d, err := time.ParseDuration(expiryStr); err == nil && d > 0 {
			expiry = d
		}
	}
	if expiry > 7*24*time.Hour {
		expiry = 7 * 24 * time.Hour
	}
	signed, err := c.s3.PresignGet(ctx.Request.Context(), path, expiry)
	if err != nil {
		request.Respond(ctx, nil, err)
		return
	}
	request.Respond(ctx, &SignedURLResponse{URL: signed}, nil)
}

func buildKey(filename, prefix string, userID int) string {
	safe := sanitizeFilename(filename)
	if safe == "" {
		safe = "file"
	}
	ts := strconv.FormatInt(time.Now().UnixNano(), 10)
	if prefix != "" {
		return prefix + "/" + strconv.Itoa(userID) + "/" + ts + "_" + safe
	}
	return strconv.Itoa(userID) + "/" + ts + "_" + safe
}

var unsafeFilename = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func sanitizeFilename(name string) string {
	base := filepath.Base(name)
	return unsafeFilename.ReplaceAllString(base, "_")
}
