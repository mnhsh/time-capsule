package storage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client *s3.Client
	bucket string
}

func NewS3Storage(ctx context.Context, bucket, region string, isLocal bool) (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if isLocal {
			// Point to LocalStack instead of actual AWS
			o.BaseEndpoint = aws.String("http://localhost:4566")
			o.UsePathStyle = true
		}
	})

	return &S3Storage{client: client, bucket: bucket}, nil
}

func (s *S3Storage) Upload(ctx context.Context, key string, body io.Reader) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	return err
}
