package s3

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/Conty111/AlfredoBot/internal/configs"
	"github.com/Conty111/AlfredoBot/internal/interfaces"
)

// S3ClientImpl implements the S3Client interface
type S3ClientImpl struct {
	client *s3.Client
}

func NewClient(cfg *configs.S3Config) (*s3.Client, error) {
	// Update endpoint to use https if UseSSL is true
	endpoint := cfg.Endpoint
	if cfg.UseSSL && strings.HasPrefix(endpoint, "http://") {
		endpoint = "https://" + strings.TrimPrefix(endpoint, "http://")
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     cfg.AccessKeyID,
				SecretAccessKey: cfg.SecretAccessKey,
			}, nil
		})),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true
		// Skip TLS verification for development environments
		if cfg.UseSSL {
			o.HTTPClient = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true, // For development only
					},
				},
			}
		}
	}), nil
}

// NewS3Client creates a new S3Client implementation
func NewS3Client(client *s3.Client) interfaces.S3Client {
	return &S3ClientImpl{
		client: client,
	}
}

// UploadFile uploads a file to S3
func (c *S3ClientImpl) UploadFile(ctx context.Context, bucket, key string, file io.Reader) error {
	// Read the entire file into memory to determine content length
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create a new reader from the bytes
	fileReader := bytes.NewReader(fileBytes)

	// Upload with content length
	_, err = c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          fileReader,
		ContentLength: aws.Int64(int64(len(fileBytes))),
	})
	return err
}

// DownloadFile downloads a file from S3
func (c *S3ClientImpl) DownloadFile(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	resp, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// DeleteFile deletes a file from S3
func (c *S3ClientImpl) DeleteFile(ctx context.Context, bucket, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// GeneratePresignedURL generates a presigned URL for an S3 object
func (c *S3ClientImpl) GeneratePresignedURL(ctx context.Context, bucket, key string, expiresIn int64) (string, error) {
	presignClient := s3.NewPresignClient(c.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expiresIn) * time.Second
	})
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
