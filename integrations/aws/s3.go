package aws

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/krsoninikhil/go-rest-kit/apperrors"
)

// Config holds S3 configuration (bucket, region).
// Credentials are loaded from the environment (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
// or the default credential chain (e.g. IAM role).
type Config struct {
	Bucket string `mapstructure:"bucket"`
	Region string `mapstructure:"region"`
}

// S3 provides PutObject, PresignGet, and PublicURL for a bucket.
type S3 struct {
	client *s3.Client
	bucket string
	region string
}

// NewS3 creates an S3 client using the default credential chain.
func NewS3(ctx context.Context, cfg Config) (*S3, error) {
	if cfg.Bucket == "" {
		return nil, apperrors.NewInvalidParamsError("s3", fmt.Errorf("bucket is required"))
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, apperrors.NewServerError(fmt.Errorf("load aws config: %w", err))
	}
	client := s3.NewFromConfig(awsCfg)
	return &S3{client: client, bucket: cfg.Bucket, region: cfg.Region}, nil
}

// PutObject uploads body to the given key.
// The public parameter is unused: ACLs are not set so the bucket works with "Bucket owner enforced"
// object ownership. To make objects public, use a bucket policy (e.g. allow GetObject on profiles/*).
func (s *S3) PutObject(ctx context.Context, key string, body io.Reader, contentType string, public bool) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	}
	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return apperrors.NewServerError(fmt.Errorf("s3 put object: %w", err))
	}
	return nil
}

// PresignGet returns a signed URL for GET on the given key.
func (s *S3) PresignGet(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", apperrors.NewInvalidParamsError("s3", fmt.Errorf("key is required"))
	}
	presigner := s3.NewPresignClient(s.client)
	req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", apperrors.NewServerError(fmt.Errorf("presign get object: %w", err))
	}
	return req.URL, nil
}

// PublicURL returns the permanent public URL for the key (virtual-hosted style).
func (s *S3) PublicURL(key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key)
}
