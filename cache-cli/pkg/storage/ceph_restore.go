package storage

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const defaultS3DownloadPartSize = int64(8 * 1024 * 1024)

// Restore downloads the cache object using aws-cli-style sequential 8 MiB
// ranged GETs (instead of the SDK's concurrent downloader). Fixed-size, aligned
// ranges let the pull-through cache cache and serve the individual chunks.
func (s *CephStorage) Restore(key string) (*os.File, error) {
	tempFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-*", key))
	if err != nil {
		return nil, err
	}

	bucketKey := fmt.Sprintf("%s/%s", s.Project, key)

	headOutput, err := s.Client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: &s.Bucket,
		Key:    &bucketKey,
	})
	if err != nil {
		_ = tempFile.Close()
		return nil, err
	}

	if headOutput.ContentLength > 0 {
		for start := int64(0); start < headOutput.ContentLength; start += defaultS3DownloadPartSize {
			end := start + defaultS3DownloadPartSize - 1
			if end >= headOutput.ContentLength {
				end = headOutput.ContentLength - 1
			}

			byteRange := fmt.Sprintf("bytes=%d-%d", start, end)
			input := &s3.GetObjectInput{
				Bucket: &s.Bucket,
				Key:    &bucketKey,
				Range:  &byteRange,
			}

			// Guard against the object changing between ranged requests.
			if headOutput.ETag != nil {
				input.IfMatch = headOutput.ETag
			}

			if err := s.copyRange(tempFile, input); err != nil {
				_ = tempFile.Close()
				return nil, err
			}
		}
	} else {
		input := &s3.GetObjectInput{Bucket: &s.Bucket, Key: &bucketKey}
		if err := s.copyRange(tempFile, input); err != nil {
			_ = tempFile.Close()
			return nil, err
		}
	}

	return tempFile, tempFile.Close()
}

func (s *CephStorage) copyRange(tempFile *os.File, input *s3.GetObjectInput) error {
	output, err := s.Client.GetObject(context.TODO(), input)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(tempFile, output.Body)
	closeErr := output.Body.Close()

	if copyErr != nil {
		return copyErr
	}

	return closeErr
}
