package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/officeryoda/dozingo/internal/config"
)

type Garage struct {
	bucketName string
	s3Client   *s3.Client
}

func New(ctx context.Context, dzgCfg *config.Config) Garage {
	cfg, err := s3cfg.LoadDefaultConfig(ctx,
		s3cfg.WithRegion("garage"), // Garage ignores the region but the SDK requires a string
		s3cfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(dzgCfg.GarageAccessKey, dzgCfg.GarageSecretKey, "")),
	)
	if err != nil {
		slog.Error("loading sdk failed", "error", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(dzgCfg.GarageEndpoint)
		o.UsePathStyle = true
	})

	return Garage{
		s3Client:   s3Client,
		bucketName: dzgCfg.GarageBucketName,
	}
}

func (g *Garage) Upload(ctx context.Context, objectKey string, img Image) error {
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

func TestGarageUpload(dzgCfg *config.Config) {
	ctx := context.Background()
	garage := New(ctx, dzgCfg)

	// Fetch random avatar
	seed := fmt.Sprint(time.Now().UnixMilli())
	img, _ := RandomProfilePictureBots(seed)
	targetFilename := fmt.Sprintf("%s%s", seed, img.Extension)

	// Upload to Garage
	err := garage.Upload(ctx, targetFilename, img)
	if err != nil {
		slog.Error("upload failed", "error", err)
	}

	// Generate the public URL that your frontend will use
	// DOZINGO: In production, change port 3902 to your public web proxy endpoint layout
	publicURL := fmt.Sprintf("http://%s.web.garage.localhost:3902/%s", dzgCfg.GarageBucketName, targetFilename)
	slog.Info("Upload successful! Access your file anonymously at ", "url", publicURL)
}
