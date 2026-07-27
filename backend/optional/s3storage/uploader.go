// Package s3storage implements storage.Storage via S3 multipart upload.
// This package lives outside the main module so AWS SDK deps are optional.
// See README.md in this directory.
package s3storage

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	S3PartSize         = 5 * 1024 * 1024
	MaxConcurrentParts = 5
	S3MaxParts         = 10000
)

// Uploader handles streaming uploads to S3.
type Uploader struct {
	client *s3.Client
	bucket string
}

// New creates a new S3 uploader. Requires AWS credentials via the default chain.
func New(bucket string) (*Uploader, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error loading AWS config: %w", err)
	}

	return &Uploader{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}, nil
}

// Save implements storage.Storage by streaming to S3 multipart upload.
func (u *Uploader) Save(ctx context.Context, name string, r io.Reader, contentType string) (string, error) {
	key := fmt.Sprintf("uploads/%s", name)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := u.UploadStream(ctx, key, r, contentType)
	if err != nil {
		return "", err
	}
	return key, nil
}

// UploadStream streams data from a reader directly to S3 using multipart upload.
func (u *Uploader) UploadStream(
	ctx context.Context,
	key string,
	reader io.Reader,
	contentType string,
) (*s3.CompleteMultipartUploadOutput, error) {
	createResp, err := u.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("error initiating multipart upload: %w", err)
	}

	uploadID := createResp.UploadId
	var completedParts []types.CompletedPart
	partNumber := int32(1)

	defer func() {
		if err != nil {
			u.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(u.bucket),
				Key:      aws.String(key),
				UploadId: uploadID,
			})
		}
	}()

	buf := make([]byte, S3PartSize)

	for {
		n, readErr := io.ReadFull(reader, buf)

		if n > 0 {
			if partNumber > S3MaxParts {
				err = fmt.Errorf("multipart upload would exceed S3's %d part limit", S3MaxParts)
				return nil, err
			}

			uploadResp, uploadErr := u.client.UploadPart(ctx, &s3.UploadPartInput{
				Bucket:        aws.String(u.bucket),
				Key:           aws.String(key),
				PartNumber:    aws.Int32(partNumber),
				UploadId:      uploadID,
				Body:          &readSeeker{data: buf[:n]},
				ContentLength: aws.Int64(int64(n)),
			})
			if uploadErr != nil {
				err = fmt.Errorf("error uploading part %d: %w", partNumber, uploadErr)
				return nil, err
			}

			completedParts = append(completedParts, types.CompletedPart{
				ETag:       uploadResp.ETag,
				PartNumber: aws.Int32(partNumber),
			})
			partNumber++
		}

		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
		if readErr != nil {
			err = fmt.Errorf("error reading data: %w", readErr)
			return nil, err
		}
	}

	if len(completedParts) == 0 {
		err = fmt.Errorf("multipart upload requires at least one part; use PutObject for empty files")
		return nil, err
	}

	completeResp, err := u.client.CompleteMultipartUpload(ctx,
		&s3.CompleteMultipartUploadInput{
			Bucket:   aws.String(u.bucket),
			Key:      aws.String(key),
			UploadId: uploadID,
			MultipartUpload: &types.CompletedMultipartUpload{
				Parts: completedParts,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("error completing multipart upload: %w", err)
	}

	return completeResp, nil
}

// UploadStreamConcurrent uploads parts to S3 concurrently.
func (u *Uploader) UploadStreamConcurrent(
	ctx context.Context,
	key string,
	reader io.Reader,
	contentType string,
) (*s3.CompleteMultipartUploadOutput, error) {
	createResp, err := u.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("error initiating multipart upload: %w", err)
	}

	uploadID := createResp.UploadId

	type partResult struct {
		partNumber int32
		etag       *string
		err        error
	}

	results := make(chan partResult, S3MaxParts)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, MaxConcurrentParts)

	partNumber := int32(1)
	var readErr error

	for readErr == nil {
		buf := make([]byte, S3PartSize)
		n, err := io.ReadFull(reader, buf)
		readErr = err

		if n > 0 {
			if partNumber > S3MaxParts {
				readErr = fmt.Errorf("multipart upload would exceed S3's %d part limit", S3MaxParts)
				break
			}

			wg.Add(1)
			semaphore <- struct{}{}

			go func(pn int32, data []byte) {
				defer wg.Done()
				defer func() { <-semaphore }()

				resp, err := u.client.UploadPart(ctx, &s3.UploadPartInput{
					Bucket:        aws.String(u.bucket),
					Key:           aws.String(key),
					PartNumber:    aws.Int32(pn),
					UploadId:      uploadID,
					Body:          &readSeeker{data: data},
					ContentLength: aws.Int64(int64(len(data))),
				})

				if err != nil {
					results <- partResult{partNumber: pn, err: err}
					return
				}
				results <- partResult{partNumber: pn, etag: resp.ETag}
			}(partNumber, buf[:n])

			partNumber++
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	partsMap := make(map[int32]*string)
	for result := range results {
		if result.err != nil {
			u.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(u.bucket),
				Key:      aws.String(key),
				UploadId: uploadID,
			})
			return nil, result.err
		}
		partsMap[result.partNumber] = result.etag
	}

	if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
		u.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(u.bucket),
			Key:      aws.String(key),
			UploadId: uploadID,
		})
		return nil, fmt.Errorf("error reading data: %w", readErr)
	}

	if len(partsMap) == 0 {
		u.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(u.bucket),
			Key:      aws.String(key),
			UploadId: uploadID,
		})
		return nil, fmt.Errorf("multipart upload requires at least one part; use PutObject for empty files")
	}

	completedParts := make([]types.CompletedPart, 0, len(partsMap))
	for i := int32(1); i < partNumber; i++ {
		completedParts = append(completedParts, types.CompletedPart{
			ETag:       partsMap[i],
			PartNumber: aws.Int32(i),
		})
	}

	completeResp, err := u.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(u.bucket),
		Key:      aws.String(key),
		UploadId: uploadID,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		u.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(u.bucket),
			Key:      aws.String(key),
			UploadId: uploadID,
		})
		return nil, fmt.Errorf("error completing multipart upload: %w", err)
	}
	return completeResp, nil
}
