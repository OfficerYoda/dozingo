// Package storage wraps the Garage S3-compatible object store used for avatars and other blobs.
package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/officeryoda/dozingo/internal/config"
)

type s3API interface {
	PutObject(ctx context.Context, in *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	ListObjectsV2(ctx context.Context, in *s3.ListObjectsV2Input, opts ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	DeleteObjects(ctx context.Context, in *s3.DeleteObjectsInput, opts ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

type Garage struct {
	bucketName string
	s3Client   s3API
}

type ObjectUploader interface {
	Upload(ctx context.Context, objectKey string, img *Image) error
}

func NewGarage(ctx context.Context, dzgCfg *config.Config) *Garage {
	cfg, err := s3cfg.LoadDefaultConfig(
		ctx,
		s3cfg.WithRegion("garage"), // Garage ignores the region but the SDK requires a string
		s3cfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(dzgCfg.GarageAccessKey, dzgCfg.GarageSecretKey, ""),
		),
	)
	if err != nil {
		slog.Error("loading sdk failed", "error", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(dzgCfg.GarageEndpoint)
		o.UsePathStyle = true
	})

	return &Garage{
		s3Client:   s3Client,
		bucketName: dzgCfg.GarageBucketName,
	}
}

func (g *Garage) Upload(ctx context.Context, objectKey string, img *Image) error {
	_, err := g.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(g.bucketName),
		Key:         aws.String(objectKey),
		Body:        img.File,
		ContentType: aws.String(img.ContentType),
	})
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}

	return nil
}
